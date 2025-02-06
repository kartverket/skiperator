package reconciliation

import (
	"context"
	"github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/jwtAuth"
	"github.com/kartverket/skiperator/pkg/log"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ObjectType string

const (
	ApplicationType ObjectType = "Application"
	JobType         ObjectType = "SKIPJob"
	NamespaceType   ObjectType = "Namespace"
	RoutingType     ObjectType = "Routing"
)

type Reconciliation interface {
	GetLogger() log.Logger
	GetCtx() context.Context //TODO: remove ctx from this interface
	IsIstioEnabled() bool
	GetSKIPObject() v1alpha1.SKIPObject
	GetType() ObjectType
	GetResources() []client.Object
	AddResource(client.Object)
	GetIdentityConfigMap() *corev1.ConfigMap
	GetRestConfig() *rest.Config
	GetAuthConfigs() *jwtAuth.AuthConfigs
}

type baseReconciliation struct {
	ctx               context.Context
	logger            log.Logger
	resources         []client.Object
	istioEnabled      bool
	restConfig        *rest.Config
	identityConfigMap *corev1.ConfigMap
	skipObject        v1alpha1.SKIPObject
	authConfigs       *jwtAuth.AuthConfigs
}

func (b *baseReconciliation) GetLogger() log.Logger {
	return b.logger
}

func (b *baseReconciliation) GetCtx() context.Context {
	return b.ctx
}

func (b *baseReconciliation) IsIstioEnabled() bool {
	return b.istioEnabled
}

func (b *baseReconciliation) GetResources() []client.Object {
	return b.resources
}

func (b *baseReconciliation) AddResource(object client.Object) {
	b.resources = append(b.resources, object)
}

func (b *baseReconciliation) GetIdentityConfigMap() *corev1.ConfigMap {
	return b.identityConfigMap
}

func (b *baseReconciliation) GetRestConfig() *rest.Config {
	return b.restConfig
}

func (b *baseReconciliation) GetSKIPObject() v1alpha1.SKIPObject {
	return b.skipObject
}

func (b *baseReconciliation) GetAuthConfigs() *jwtAuth.AuthConfigs {
	return b.authConfigs
}
