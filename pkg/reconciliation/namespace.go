package reconciliation

import (
	"context"
	"github.com/kartverket/skiperator/pkg/log"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type NamespaceReconciliation struct {
	ctx               context.Context
	namespace         *corev1.Namespace
	logger            log.Logger
	resources         []*client.Object
	istioEnabled      bool
	restConfig        *rest.Config
	identityConfigMap *corev1.ConfigMap
}

func NewNamespaceReconciliation(ctx context.Context, namespace *corev1.Namespace, logger log.Logger, restConfig *rest.Config, identityConfigMap *corev1.ConfigMap) *NamespaceReconciliation {
	return &NamespaceReconciliation{
		ctx:               ctx,
		namespace:         namespace,
		logger:            logger,
		restConfig:        restConfig,
		identityConfigMap: identityConfigMap,
	}
}

func (r *NamespaceReconciliation) GetLogger() log.Logger {
	return r.logger
}

func (r *NamespaceReconciliation) GetCtx() context.Context {
	return r.ctx
}

func (r *NamespaceReconciliation) IsIstioEnabled() bool {
	return r.istioEnabled
}

func (r *NamespaceReconciliation) GetReconciliationObject() client.Object {
	return r.namespace
}

func (r *NamespaceReconciliation) GetType() ReconciliationObjectType {
	return NamespaceType
}

func (r *NamespaceReconciliation) GetRestConfig() *rest.Config {
	return r.restConfig
}

func (r *NamespaceReconciliation) AddResource(object *client.Object) {
	r.resources = append(r.resources, object)
}

func (r *NamespaceReconciliation) GetResources() []*client.Object {
	return r.resources
}

func (r *NamespaceReconciliation) GetCommonSpec() *CommonType {
	panic("implement me")
}

func (r *NamespaceReconciliation) GetIdentityConfigMap() *corev1.ConfigMap {
	return r.identityConfigMap
}
