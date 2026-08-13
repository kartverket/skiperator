package reconciliation

import (
	"context"

	"github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/internal/config"
	"github.com/kartverket/skiperator/pkg/auth"
	"github.com/kartverket/skiperator/pkg/log"
	"github.com/kartverket/skiperator/pkg/mesh"
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
	IsSidecarEnabled() bool
	MeshMode() mesh.Mode
	GetSKIPObject() v1alpha1.SKIPObject
	GetType() ObjectType
	GetResources() []client.Object
	AddResource(client.Object)
	GetRestConfig() *rest.Config
	GetAuthConfigs() *auth.AuthConfigs
	GetSkiperatorConfig() config.SkiperatorConfig
	GenerateLegacyRouting() bool
	SetGenerateLegacyRouting(bool)
}

type baseReconciliation struct {
	ctx                   context.Context
	logger                log.Logger
	resources             []client.Object
	meshMode              mesh.Mode
	restConfig            *rest.Config
	skipObject            v1alpha1.SKIPObject
	authConfigs           *auth.AuthConfigs
	skiperatorConfig      config.SkiperatorConfig
	generateLegacyRouting bool
}

func (b *baseReconciliation) GetLogger() log.Logger {
	return b.logger
}

func (b *baseReconciliation) GetCtx() context.Context {
	return b.ctx
}

// IsSidecarEnabled reports whether pods get an Istio sidecar. Sidecar-specific
// resources such as the Sidecar resource, the proxy metrics port, and the
// scrape network policy depend on this, and ambient namespaces must not get
// them.
func (b *baseReconciliation) IsSidecarEnabled() bool {
	return b.meshMode == mesh.ModeSidecar
}

// MeshMode exposes the mode itself, for generators that branch on more than
// the sidecar case.
func (b *baseReconciliation) MeshMode() mesh.Mode {
	return b.meshMode
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

func (b *baseReconciliation) GenerateLegacyRouting() bool {
	return b.generateLegacyRouting
}

func (b *baseReconciliation) SetGenerateLegacyRouting(generate bool) {
	b.generateLegacyRouting = generate
}
