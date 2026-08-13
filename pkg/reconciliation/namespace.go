package reconciliation

import (
	"context"

	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/log"
	"github.com/kartverket/skiperator/pkg/mesh"
	"k8s.io/client-go/rest"
)

type NamespaceReconciliation struct {
	baseReconciliation
}

func NewNamespaceReconciliation(ctx context.Context, namespace skiperatorv1alpha1.SKIPObject,
	logger log.Logger, meshMode mesh.Mode,
	restConfig *rest.Config) *NamespaceReconciliation {
	return &NamespaceReconciliation{
		baseReconciliation: baseReconciliation{
			ctx:        ctx,
			logger:     logger,
			meshMode:   meshMode,
			restConfig: restConfig,
			skipObject: namespace,
		},
	}
}

func (r *NamespaceReconciliation) GetType() ObjectType {
	return NamespaceType
}
