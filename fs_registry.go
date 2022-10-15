package main

import (
	iofs "io/fs"
	"path"
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
	var results []Package
	for _, p := range packages {
		if !p.IsDir() {
			continue
		}
		includeAllPackages := req.Package == nil
		includeAllVersions := includeAllPackages || req.Package.Version == nil

		if includeAllPackages || req.Package.Name == p.Name() {
			pkg := &Package{
				Name: p.Name(),
			}
			versions, err := iofs.ReadDir(r.fs, path.Join(r.root, p.Name()))
			if err != nil {
				return nil, err
			}
			for _, v := range versions {
				if !v.IsDir() {
					continue
				}
				if includeAllVersions || req.Package.Version.Expression == v.Name() {
					ver := &Version{
						Number: v.Name(),
					}
					pkg.Versions = append(pkg.Versions, *ver)
				}
			}
			results = append(results, *pkg)
		}
	}
	return results, nil
}
