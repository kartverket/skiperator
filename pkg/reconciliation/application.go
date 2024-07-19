package reconciliation

import (
	"context"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/log"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ApplicationReconciliation struct {
	ctx               context.Context
	application       *skiperatorv1alpha1.Application
	logger            log.Logger
	resources         []*client.Object
	istioEnabled      bool
	restConfig        *rest.Config
	identityConfigMap *corev1.ConfigMap
}

func NewApplicationReconciliation(ctx context.Context, application *skiperatorv1alpha1.Application, logger log.Logger, istioEnabled bool, restConfig *rest.Config, identityConfigMap *corev1.ConfigMap) *ApplicationReconciliation {
	return &ApplicationReconciliation{
		ctx:               ctx,
		application:       application,
		logger:            logger,
		istioEnabled:      istioEnabled,
		restConfig:        restConfig,
		identityConfigMap: identityConfigMap,
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

func (r *ApplicationReconciliation) GetRestConfig() *rest.Config {
	return r.restConfig
}

func (r *ApplicationReconciliation) AddResource(object *client.Object) {
	r.resources = append(r.resources, object)
}

func (r *ApplicationReconciliation) GetResources() []*client.Object {
	return r.resources
}

func (r *ApplicationReconciliation) GetCommonSpec() *CommonType {
	return &CommonType{
		GCP:          r.application.Spec.GCP,
		AccessPolicy: r.application.Spec.AccessPolicy,
	}
}

func (r *ApplicationReconciliation) GetIdentityConfigMap() *corev1.ConfigMap {
	return r.identityConfigMap
}
