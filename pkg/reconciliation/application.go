package reconciliation

import (
	"context"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/log"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ApplicationReconciliation struct {
	ctx          context.Context
	application  *skiperatorv1alpha1.Application
	logger       log.Logger
	objects      []*client.Object
	istioEnabled bool
	restConfig   *rest.Config
}

func NewApplicationReconciliation(ctx context.Context, application *skiperatorv1alpha1.Application, logger log.Logger, restConfig *rest.Config) *ApplicationReconciliation {
	return &ApplicationReconciliation{
		ctx:         ctx,
		application: application,
		logger:      logger,
		restConfig:  restConfig,
	}
}

func (r *ApplicationReconciliation) GetLogger() log.Logger {
	return r.logger
}

func (r *ApplicationReconciliation) GetCtx() context.Context {
	return r.ctx
}

func (r *ApplicationReconciliation) IsIstioEnabled() bool {
	return r.istioEnabled
}

func (r *ApplicationReconciliation) GetReconciliationObject() client.Object {
	return r.application
}

func (r *ApplicationReconciliation) GetType() ReconciliationObjectType {
	return ApplicationType
}

func (r *ApplicationReconciliation) AddSyncObject(object *client.Object) {
	r.objects = append(r.objects, object)
}

func (r *ApplicationReconciliation) GetSyncObjects() []*client.Object {
	return r.objects
}
