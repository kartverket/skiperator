package reconciliation

import (
	"context"

	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/log"
	"k8s.io/client-go/rest"
)

type RoutingReconciliation struct {
	baseReconciliation
}

func NewRoutingReconciliation(ctx context.Context, routing *skiperatorv1alpha1.Routing,
	logger log.Logger, istioEnabled bool, restConfig *rest.Config) *RoutingReconciliation {
	return &RoutingReconciliation{
		baseReconciliation: baseReconciliation{
			ctx:          ctx,
			logger:       logger,
			istioEnabled: istioEnabled,
			restConfig:   restConfig,
			skipObject:   routing,
		},
	}
}

func (r *RoutingReconciliation) GetType() ObjectType {
	return RoutingType
}
