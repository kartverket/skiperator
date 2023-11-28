//go:build tools

package skiperator

import (
	_ "go.etcd.io/etcd/server/v3"
	_ "sigs.k8s.io/controller-tools/cmd/controller-gen"
)
