package reconciliation

import (
	"context"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/v3/api/v1alpha1"
	"github.com/kartverket/skiperator/v3/pkg/auth"
	"github.com/kartverket/skiperator/v3/pkg/log"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
)

type ApplicationReconciliation struct {
	baseReconciliation
}

func NewApplicationReconciliation(ctx context.Context, application *skiperatorv1alpha1.Application,
	logger log.Logger, istioEnabled bool, restConfig *rest.Config,
	identityConfigMap *corev1.ConfigMap, authConfigs *auth.AuthConfigs) *ApplicationReconciliation {
	return &ApplicationReconciliation{
		baseReconciliation: baseReconciliation{
			ctx:               ctx,
			logger:            logger,
			istioEnabled:      istioEnabled,
			restConfig:        restConfig,
			identityConfigMap: identityConfigMap,
			skipObject:        application,
			authConfigs:       authConfigs,
		},
	}
}

func (r *ApplicationReconciliation) GetType() ObjectType {
	return ApplicationType
}
