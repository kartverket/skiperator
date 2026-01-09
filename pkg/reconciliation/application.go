package reconciliation

import (
	"context"

	skiperatorv1alpha1 "github.com/kartverket/skiperator/v2/api/v1alpha1"
	"github.com/kartverket/skiperator/v2/internal/config"
	"github.com/kartverket/skiperator/v2/pkg/auth"
	"github.com/kartverket/skiperator/v2/pkg/log"
	"k8s.io/client-go/rest"
)

type ApplicationReconciliation struct {
	baseReconciliation
}

func NewApplicationReconciliation(ctx context.Context, application *skiperatorv1alpha1.Application,
	logger log.Logger, istioEnabled bool, restConfig *rest.Config, authConfigs *auth.AuthConfigs, skiperatorConfig config.SkiperatorConfig) *ApplicationReconciliation {
	return &ApplicationReconciliation{
		baseReconciliation: baseReconciliation{
			ctx:              ctx,
			logger:           logger,
			istioEnabled:     istioEnabled,
			restConfig:       restConfig,
			skipObject:       application,
			authConfigs:      authConfigs,
			skiperatorConfig: skiperatorConfig,
		},
	}
}

func (r *ApplicationReconciliation) GetType() ObjectType {
	return ApplicationType
}
