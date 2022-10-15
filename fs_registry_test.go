package main_test

import (
	iofs "io/fs"
	"testing/fstest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	verman "github.com/patrickhuber/go-verman"
)

var _ = Describe("FsRegistry", func() {
	var (
		fs  iofs.FS
		reg verman.Registry
	)
	BeforeEach(func() {
		fs = fstest.MapFS{
			"dog/1.0.0/file.txt": {
				Data: []byte("woof"),
			},
			"dog/1.0.1/file.txt": {
				Data: []byte("woof!"),
			},
			"dog/2.0.0/file.txt": {
				Data: []byte("woof woof"),
			},
			"cat/latest": {
				Data: []byte("1.0.0"),
			},
			"cat/1.0.0/file.txt": {
				Data: []byte("meow"),
			},
			"cat/2.0.0/file.txt": {
				Data: []byte("mew2"),
			},
		}
		reg = verman.NewFsRegistry(fs, ".")
	})
	It("can list", func() {
		resp, err := reg.List(&verman.ListRequest{})
		Expect(err).To(BeNil())
		Expect(len(resp)).Should(BeNumerically(">=", 1))
	})
	It("can list package by name", func() {
		resp, err := reg.List(
			&verman.ListRequest{
				Package: &verman.PackageCriteria{
					Name: "cat"}})
		Expect(err).To(BeNil())
		Expect(len(resp)).Should(Equal(1))
	})
	It("can list package versions", func() {
		resp, err := reg.List(
			&verman.ListRequest{
				Package: &verman.PackageCriteria{
					Name: "cat",
				},
			},
		)
		Expect(err).To(BeNil())
		Expect(len(resp)).To(Equal(1))
		for _, p := range resp {
			Expect(len(p.Versions)).To(Equal(2))
		}
	})
	It("can list specific package version", func() {
		resp, err := reg.List(
			&verman.ListRequest{
				Package: &verman.PackageCriteria{
					Name: "cat",
					Version: &verman.PackageVersionCriteria{
						Expression: "1.0.0",
					},
				},
			},
		)

		Expect(err).To(BeNil())
		Expect(len(resp)).To(Equal(1))
		for _, p := range resp {
			Expect(len(p.Versions)).To(Equal(1))
			Expect(p.Versions[0].Number).To(Equal("1.0.0"))
		}
	})
	It("can list specific package with constraint", func() {
		resp, err := reg.List(
			&verman.ListRequest{
				Package: &verman.PackageCriteria{
					Name: "cat",
					Version: &verman.PackageVersionCriteria{
						Expression: "=2.0.0",
					},
				},
			},
		)

		Expect(err).To(BeNil())
		Expect(len(resp)).To(Equal(1))
		for _, p := range resp {
			Expect(len(p.Versions)).To(Equal(1))
			Expect(p.Versions[0].Number).To(Equal("2.0.0"))
		}
	})
	When("latest file present", func() {
		It("can list latest package version", func() {
			resp, err := reg.List(
				&verman.ListRequest{
					Package: &verman.PackageCriteria{
						Name: "cat",
						Version: &verman.PackageVersionCriteria{
							Latest: true,
						},
					},
				},
			)

			Expect(err).To(BeNil())
			Expect(len(resp)).To(Equal(1))
			Expect(len(resp[0].Versions)).To(Equal(1))
			Expect(resp[0].Versions[0].Number).To(Equal("1.0.0"))
		})
	})
	When("latest file missing", func() {
		It("can list latest package version", func() {
			resp, err := reg.List(
				&verman.ListRequest{
					Package: &verman.PackageCriteria{
						Name: "dog",
						Version: &verman.PackageVersionCriteria{
							Latest: true,
						},
					},
				},
			)

			Expect(err).To(BeNil())
			Expect(len(resp)).To(Equal(1))
			Expect(len(resp[0].Versions)).To(Equal(1))
			Expect(resp[0].Versions[0].Number).To(Equal("2.0.0"))
		})
	})
})
