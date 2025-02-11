package digdirator

import (
	"fmt"
	"github.com/kartverket/skiperator/api/v1alpha1/istiotypes"
	"github.com/nais/digdirator/pkg/secrets"
	nais_io_v1 "github.com/nais/liberator/pkg/apis/nais.io/v1"
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
	Client *nais_io_v1.MaskinportenClient
	Error  error
}

const MaskinPortenName = "maskinporten"

func (i *Maskinporten) IsEnabled() bool {
	return i.Enabled && i.Authentication.Enabled
}

func (i *Maskinporten) GetDigdiratorName() DigdiratorName {
	return MaskinPortenName
}

func (i *Maskinporten) GetProvidedSecretName() *string {
	return i.Authentication.SecretName
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

func (i *Maskinporten) GetDigdiratorClientOwnerRef(digdiratorClients DigdiratorClients) (*[]v1.OwnerReference, error) {
	err := i.HandleDigdiratorClientError(digdiratorClients)
	if err == nil {
		return nil, err
	}
	return &digdiratorClients.MaskinPortenClient.Client.OwnerReferences, nil
}

func (i *Maskinporten) GetGeneratedDigdiratorSecret(digdiratorClients DigdiratorClients) (*string, error) {
	err := i.HandleDigdiratorClientError(digdiratorClients)
	if err == nil {
		return nil, err
	}
	return &digdiratorClients.MaskinPortenClient.Client.Spec.SecretName, nil
}

func (i *Maskinporten) HandleDigdiratorClientError(digdiratorClients DigdiratorClients) error {
	if digdiratorClients.MaskinPortenClient.Error != nil {
		return fmt.Errorf("failed to get MaskinportenClient: %w", digdiratorClients.MaskinPortenClient.Error)
	}
	if digdiratorClients.MaskinPortenClient.Client == nil {
		return fmt.Errorf("MaskinportenClient not found")
	}
	return nil
}
