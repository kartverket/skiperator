package reconciliation

import (
	"context"

	"github.com/kartverket/skiperator/v2/api/v1alpha1"
	"github.com/kartverket/skiperator/v2/internal/config"
	"github.com/kartverket/skiperator/v2/pkg/auth"
	"github.com/kartverket/skiperator/v2/pkg/log"
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
	GetRestConfig() *rest.Config
	GetAuthConfigs() *auth.AuthConfigs
	GetSkiperatorConfig() config.SkiperatorConfig
}

type baseReconciliation struct {
	ctx              context.Context
	logger           log.Logger
	resources        []client.Object
	istioEnabled     bool
	restConfig       *rest.Config
	skipObject       v1alpha1.SKIPObject
	authConfigs      *auth.AuthConfigs
	skiperatorConfig config.SkiperatorConfig
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

func (b *baseReconciliation) GetRestConfig() *rest.Config {
	return b.restConfig
}

func (b *baseReconciliation) GetSKIPObject() v1alpha1.SKIPObject {
	return b.skipObject
}

func (b *baseReconciliation) GetAuthConfigs() *auth.AuthConfigs {
	return b.authConfigs
}

func (b *baseReconciliation) GetSkiperatorConfig() config.SkiperatorConfig {
	return b.skiperatorConfig
}
