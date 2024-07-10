package reconciliation

import (
	"context"
	"github.com/kartverket/skiperator/pkg/log"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type NamespaceReconciliation struct {
	ctx          context.Context
	namespace    *v1.Namespace
	logger       log.Logger
	resources    []*client.Object
	istioEnabled bool
	restConfig   *rest.Config
}

func NewNamespaceReconciliation(ctx context.Context, namespace *v1.Namespace, logger log.Logger, restConfig *rest.Config) *NamespaceReconciliation {
	return &NamespaceReconciliation{
		ctx:        ctx,
		namespace:  namespace,
		logger:     logger,
		restConfig: restConfig,
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

func (r *NamespaceReconciliation) AddResource(object *client.Object) {
	r.resources = append(r.resources, object)
}

func (r *NamespaceReconciliation) GetResources() []*client.Object {
	return r.resources
}
