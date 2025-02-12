package digdirator

import (
	nais_io_v1 "github.com/nais/liberator/pkg/apis/nais.io/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type DigdiratorName string

type DigdiratorURIs struct {
	Name      DigdiratorName
	IssuerURI string
	JwksURI   string
	ClientID  string
}

type DigdiratorProvider interface {
	IsEnabled() bool
	GetDigdiratorName() DigdiratorName
	GetProvidedSecretName() *string
	GetIgnoredPaths() []string
	GetIssuerKey() string
	GetJwksKey() string
	GetClientIDKey() string
}

type DigdiratorClient interface {
	GetOwnerReferences() []v1.OwnerReference
	GetSecretName() string
}

type MaskinportenClient struct {
	nais_io_v1.MaskinportenClient
}

func (m *MaskinportenClient) GetSecretName() string {
	return m.Spec.SecretName
}

type IDPortenClient struct {
	nais_io_v1.IDPortenClient
}

func (i *IDPortenClient) GetSecretName() string {
	return i.Spec.SecretName
}
