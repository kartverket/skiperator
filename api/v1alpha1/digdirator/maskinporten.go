package digdirator

import (
	"github.com/kartverket/skiperator/api/v1alpha1/istiotypes"
	"github.com/nais/digdirator/pkg/secrets"
	nais_io_v1 "github.com/nais/liberator/pkg/apis/nais.io/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// https://github.com/nais/liberator/blob/c9da4cf48a52c9594afc8a4325ff49bbd359d9d2/pkg/apis/nais.io/v1/naiserator_types.go#L376
//
// +kubebuilder:object:generate=true
type Maskinporten struct {
	// The name of the Client as shown in Digitaliseringsdirektoratet's Samarbeidsportal
	// Meant to be a human-readable name for separating clients in the portal
	ClientName *string `json:"clientName,omitempty"`

	// If enabled, provisions and configures a Maskinporten client with consumed scopes and/or Exposed scopes with DigDir.
	Enabled bool `json:"enabled"`

	// Schema to configure Maskinporten clients with consumed scopes and/or exposed scopes.
	Scopes *nais_io_v1.MaskinportenScope `json:"scopes,omitempty"`

	// Authentication specifies how incoming JWT's should be validated.
	Authentication *istiotypes.Authentication `json:"authentication,omitempty"`
}

type MaskinportenClient struct {
	Client nais_io_v1.MaskinportenClient
}

func (m *MaskinportenClient) GetOwnerReferences() []v1.OwnerReference {
	return m.Client.GetOwnerReferences()
}

func (m *MaskinportenClient) GetSecretName() string {
	return m.Client.Spec.SecretName
}

const MaskinPortenName = "maskinporten"

func (i *Maskinporten) IsEnabled() bool {
	return i != nil && i.Authentication != nil && i.Authentication.Enabled
}

func (i *Maskinporten) GetAuthSpec() istiotypes.Authentication {
	return *i.Authentication
}

func (i *Maskinporten) GetDigdiratorName() DigdiratorName {
	return MaskinPortenName
}

func (i *Maskinporten) GetProvidedSecretName() *string {
	return i.Authentication.SecretName
}

func (i *Maskinporten) GetPaths() []string {
	var paths []string
	if i.IsEnabled() {
		if i.Authentication.Paths != nil {
			paths = append(paths, *i.Authentication.Paths...)
		}
	}
	return paths
}

func (i *Maskinporten) GetIgnoredPaths() []string {
	var ignoredPaths []string
	if i.IsEnabled() {
		if i.Authentication.IgnorePaths != nil {
			ignoredPaths = append(ignoredPaths, *i.Authentication.IgnorePaths...)
		}
	}
	return ignoredPaths
}

func (i *Maskinporten) GetIssuerKey() string {
	return secrets.MaskinportenIssuerKey
}

func (i *Maskinporten) GetJwksKey() string {
	return secrets.MaskinportenJwksUriKey
}

func (i *Maskinporten) GetClientIDKey() string {
	return secrets.MaskinportenClientIDKey
}
