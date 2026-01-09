package reconciliation

import (
	"context"

	skiperatorv1alpha1 "github.com/kartverket/skiperator/v2/api/v1alpha1"
	"github.com/kartverket/skiperator/v2/pkg/log"
	"k8s.io/client-go/rest"
)

type NamespaceReconciliation struct {
	baseReconciliation
}

func NewNamespaceReconciliation(ctx context.Context, namespace skiperatorv1alpha1.SKIPObject,
	logger log.Logger, istioEnabled bool,
	restConfig *rest.Config) *NamespaceReconciliation {
	return &NamespaceReconciliation{
		baseReconciliation: baseReconciliation{
			ctx:          ctx,
			logger:       logger,
			istioEnabled: istioEnabled,
			restConfig:   restConfig,
			skipObject:   namespace,
		},
	}
}

func (r *NamespaceReconciliation) GetType() ObjectType {
	return NamespaceType
}
