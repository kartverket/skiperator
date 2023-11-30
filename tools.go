//go:build tools

package skiperator

import (
	_ "github.com/kyverno/chainsaw"
	_ "sigs.k8s.io/controller-tools/cmd/controller-gen"
)
