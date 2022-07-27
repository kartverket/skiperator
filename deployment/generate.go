package deployment

//go:generate controller-gen crd paths=../api/... output:dir=.
//go:generate controller-gen rbac:roleName=skiperator paths=../... output:dir=.
