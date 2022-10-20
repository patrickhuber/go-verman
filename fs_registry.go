package main

import (
	"errors"
	iofs "io/fs"
	"net/url"
	"os"
	"path"
	"sort"

	"github.com/Masterminds/semver/v3"
)

func NewFsRegistry(fs iofs.FS, root string) Registry {
	return &fsRegistry{
		fs:   fs,
		root: root,
	}
}

type fsRegistry struct {
	fs   iofs.FS
	root string
}

func (r *fsRegistry) List(req *ListRequest) ([]Package, error) {
	packages, err := iofs.ReadDir(r.fs, r.root)
	if err != nil {
		return nil, err
	}

	// if no criteria specified, all packages are returned
	includeAllPackages := req.Package == nil

	// if no version criteria are specfied, all versions are returned
	includeAllVersions := includeAllPackages ||
		req.Package.Version == nil ||
		req.Package.Version.Expression == ""

	// if only the latest version is requested, the latest file is checked
	// if no latest file exists, the versions are sorted and the greatest is returned
	includeLatestVersion := req.Package != nil &&
		req.Package.Version != nil &&
		req.Package.Version.Latest

	var constraint *semver.Constraints
	if !includeAllVersions && !includeLatestVersion {
		constraint, err = semver.NewConstraint(req.Package.Version.Expression)
		if err != nil {
			// Handle constraint not being parsable.
			return nil, err
		}
	}

	var results []Package
	for _, p := range packages {
		if !p.IsDir() {
			continue
		}

		if !includeAllPackages && req.Package.Name != p.Name() {
			continue
		}

		pkg := &Package{
			Name: p.Name(),
		}

		versions, err := iofs.ReadDir(r.fs, path.Join(r.root, p.Name()))
		if err != nil {
			return nil, err
		}

		// if including the latest version and the latest file exists, set the constraint
		if includeLatestVersion {
			latest, err := iofs.ReadFile(r.fs, path.Join(r.root, p.Name(), "latest"))
			if err != nil && !errors.Is(err, os.ErrNotExist) {
				return nil, err
			}
			if err == nil {
				constraint, err = semver.NewConstraint(string(latest))
				if err != nil {
					return nil, err
				}
			}
		}

		semverVersions := []*semver.Version{}
		for _, v := range versions {

			// only process directories
			if !v.IsDir() {
				continue
			}

			semverVersion, err := semver.NewVersion(v.Name())
			if err != nil {
				// skip directories that are not parsable
				continue
			}
			if !includeAllVersions && constraint != nil && !constraint.Check(semverVersion) {
				continue
			}
			semverVersions = append(semverVersions, semverVersion)
		}

		// if there are no versions, continue
		if len(semverVersions) == 0 {
			continue
		}

		// sort the results if the constraint is null and latest was specified
		if constraint == nil && includeLatestVersion {
			// sort the slice and set the return value to the latest
			sort.Sort(semver.Collection(semverVersions))
			semverVersions = semverVersions[len(semverVersions)-1:]
		}
		for _, semverVersion := range semverVersions {
			if constraint != nil {
				if !constraint.Check(semverVersion) {
					continue
				}
			}
			version := &Version{
				Number: semverVersion.Original(),
			}
			pkg.Versions = append(pkg.Versions, *version)
		}
		results = append(results, *pkg)
	}
	return results, nil
}

func (r *fsRegistry) Get(req *GetRequest) (Package, error) {
	versionPath := path.Join(r.root, req.PackageName, req.PackageVersion)
	_, err := iofs.Stat(r.fs, versionPath)
	if err != nil {
		return Package{}, err
	}

	entries, err := iofs.ReadDir(r.fs, versionPath)
	if err != nil {
		return Package{}, err
	}

	files := []File{}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		ustring, err := url.JoinPath("file://", versionPath, e.Name())
		if err != nil {
			return Package{}, err
		}
		u, err := url.Parse(ustring)
		if err != nil {
			return Package{}, err
		}
		files = append(files, File{
			Name: e.Name(),
			Link: u,
		})
	}
	return Package{
			Name: req.PackageName,
			Versions: []Version{
				{
					Number: req.PackageVersion,
					Files:  files,
				}}},
		nil
}
