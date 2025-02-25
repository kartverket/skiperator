package reconciliation

import (
	"context"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/auth"
	"github.com/kartverket/skiperator/pkg/log"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
)

type ApplicationReconciliation struct {
	baseReconciliation
}

func NewApplicationReconciliation(ctx context.Context, application *skiperatorv1alpha1.Application,
	logger log.Logger, istioEnabled bool, restConfig *rest.Config,
	identityConfigMap *corev1.ConfigMap, requestAuthConfigs *auth.RequestAuthConfigs, autoLoginConfig *auth.AutoLoginConfig) *ApplicationReconciliation {
	return &ApplicationReconciliation{
		baseReconciliation: baseReconciliation{
			ctx:                ctx,
			logger:             logger,
			istioEnabled:       istioEnabled,
			restConfig:         restConfig,
			identityConfigMap:  identityConfigMap,
			skipObject:         application,
			requestAuthConfigs: requestAuthConfigs,
			autoLoginConfig:    autoLoginConfig,
		},
	}
}

func (r *ApplicationReconciliation) GetType() ObjectType {
	return ApplicationType
}
