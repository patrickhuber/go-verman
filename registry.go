package main

type Registry interface {
	List(req *ListRequest) ([]Package, error)
	Get(req *GetRequest) (Package, error)
}

type ListRequest struct {
	Package *PackageCriteria
}

type GetRequest struct {
	PackageName    string
	PackageVersion string
}

type PackageCriteria struct {
	Name    string
	Version *PackageVersionCriteria
}

type PackageVersionCriteria struct {
	Expression string
	Latest     bool
}
