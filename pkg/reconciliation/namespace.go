package reconciliation

import (
	"context"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/log"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
)

type NamespaceReconciliation struct {
	baseReconciliation
}

func NewNamespaceReconciliation(ctx context.Context, namespace skiperatorv1alpha1.SKIPObject,
	logger log.Logger, istioEnabled bool,
	restConfig *rest.Config, identityConfigMap *corev1.ConfigMap) *NamespaceReconciliation {
	return &NamespaceReconciliation{
		baseReconciliation: baseReconciliation{
			ctx:               ctx,
			logger:            logger,
			istioEnabled:      istioEnabled,
			restConfig:        restConfig,
			identityConfigMap: identityConfigMap,
			skipObject:        namespace,
		},
	}
}

func (r *NamespaceReconciliation) GetType() ObjectType {
	return NamespaceType
}
