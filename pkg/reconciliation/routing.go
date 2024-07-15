package reconciliation

import (
	"context"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/log"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type RoutingReconciliation struct {
	ctx               context.Context
	application       *skiperatorv1alpha1.Routing
	logger            log.Logger
	resources         []*client.Object
	istioEnabled      bool
	restConfig        *rest.Config
	identityConfigMap *corev1.ConfigMap
}

func NewRoutingReconciliation(ctx context.Context, routing *skiperatorv1alpha1.Routing, logger log.Logger, restConfig *rest.Config, identityConfigMap *corev1.ConfigMap) *RoutingReconciliation {
	return &RoutingReconciliation{
		ctx:               ctx,
		application:       routing,
		logger:            logger,
		restConfig:        restConfig,
		identityConfigMap: identityConfigMap,
	}
}

func (r *RoutingReconciliation) GetLogger() log.Logger {
	return r.logger
}

func (r *RoutingReconciliation) GetCtx() context.Context {
	return r.ctx
}

func (r *RoutingReconciliation) IsIstioEnabled() bool {
	return r.istioEnabled
}

func (r *RoutingReconciliation) GetReconciliationObject() client.Object {
	return r.application
}

func (r *RoutingReconciliation) GetType() ReconciliationObjectType {
	return RoutingType
}

func (r *RoutingReconciliation) GetRestConfig() *rest.Config {
	return r.restConfig
}

func (r *RoutingReconciliation) AddResource(object *client.Object) {
	r.resources = append(r.resources, object)
}

func (r *RoutingReconciliation) GetResources() []*client.Object {
	return r.resources
}

func (r *RoutingReconciliation) GetCommonSpec() *CommonType {
	panic("implement me")
}

func (r *RoutingReconciliation) GetIdentityConfigMap() *corev1.ConfigMap {
	return r.identityConfigMap
}
