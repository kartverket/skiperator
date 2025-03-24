package digdirator

import (
	"github.com/kartverket/skiperator/api/v1alpha1/istiotypes"
	"github.com/nais/digdirator/pkg/secrets"
	nais_io_v1 "github.com/nais/liberator/pkg/apis/nais.io/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const MaskinPortenName = "maskinporten"

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

	// RequestAuth specifies how incoming JWTs should be validated.
	RequestAuth *istiotypes.RequestAuth `json:"requestAuth,omitempty"`
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

func (i *Maskinporten) IsRequestAuthEnabled() bool {
	return i != nil && i.RequestAuth != nil && i.RequestAuth.Enabled
}

func (i *Maskinporten) GetAuthSpec() *istiotypes.RequestAuth {
	if i != nil && i.RequestAuth != nil {
		return i.RequestAuth
	}
	return nil
}

func (i *Maskinporten) GetDigdiratorName() DigdiratorName {
	return MaskinPortenName
}

func (i *Maskinporten) GetProvidedSecretName() *string {
	if i != nil && i.RequestAuth != nil {
		return i.RequestAuth.SecretName
	}
	return nil
}

func (i *Maskinporten) GetAuthRules() istiotypes.RequestAuthRules {
	var requestAuthRules istiotypes.RequestAuthRules
	if i.IsRequestAuthEnabled() {
		if i.RequestAuth.AuthRules != nil {
			requestAuthRules = append(requestAuthRules, *i.RequestAuth.AuthRules...)
		}
	}
	return requestAuthRules
}

func (i *Maskinporten) GetIgnoredAuthRules() istiotypes.RequestMatchers {
	var ignoredAuthRules istiotypes.RequestMatchers
	if i.IsRequestAuthEnabled() {
		if i.RequestAuth.IgnoreAuth != nil {
			ignoredAuthRules = append(ignoredAuthRules, *i.RequestAuth.IgnoreAuth...)
		}
	}
	return ignoredAuthRules
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

func (i *Maskinporten) GetTokenLocation() string {
	if i != nil && i.RequestAuth != nil && i.RequestAuth.TokenLocation != nil {
		return *i.RequestAuth.TokenLocation
	}
	return "header"
}

func (i *Maskinporten) GetAcceptedResources() []string {
	if i.IsRequestAuthEnabled() && i.RequestAuth.AcceptedResources != nil {
		return i.RequestAuth.AcceptedResources
	}
	return []string{}
}

func (i *Maskinporten) IncludesInternalTraffic() bool {
	return i.IsRequestAuthEnabled() && i.RequestAuth.IncludeInternalTraffic
}
