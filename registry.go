package main

type Registry interface {
	List(req *ListRequest) ([]Package, error)
}

type ListRequest struct {
	Package *PackageCriteria
}

type PackageCriteria struct {
	Name    string
	Version *PackageVersionCriteria
}

type PackageVersionCriteria struct {
	Expression string
	Latest     bool
}
