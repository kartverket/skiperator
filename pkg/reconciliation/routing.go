package reconciliation

import (
	"context"

	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/log"
	"github.com/kartverket/skiperator/pkg/mesh"
	"k8s.io/client-go/rest"
)

type RoutingReconciliation struct {
	baseReconciliation
	targetAppPorts map[string]int32
}

func NewRoutingReconciliation(ctx context.Context, routing *skiperatorv1alpha1.Routing,
	logger log.Logger, meshMode mesh.Mode, restConfig *rest.Config, targetAppPorts map[string]int32) *RoutingReconciliation {
	return &RoutingReconciliation{
		baseReconciliation: baseReconciliation{
			ctx:                   ctx,
			logger:                logger,
			meshMode:              meshMode,
			restConfig:            restConfig,
			skipObject:            routing,
			generateLegacyRouting: true,
		},
		targetAppPorts: targetAppPorts,
	}
}

func (r *RoutingReconciliation) GetType() ObjectType {
	return RoutingType
}

func (r *RoutingReconciliation) GetTargetAppPort(targetApp string, fallbackPort int32) int32 {
	if r.targetAppPorts == nil {
		return fallbackPort
	}
	if port, ok := r.targetAppPorts[targetApp]; ok {
		return port
	}
	return fallbackPort
}
