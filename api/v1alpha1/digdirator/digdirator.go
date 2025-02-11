package digdirator

import (
	"github.com/kartverket/skiperator/api/v1alpha1/istiotypes"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type DigdiratorName string

type DigdiratorURIs struct {
	Name      DigdiratorName
	IssuerURI string
	JwksURI   string
	ClientID  string
}

type DigdiratorClients struct {
	IdPortenClient     IdPortenClient
	MaskinPortenClient MaskinportenClient
}

type DigdiratorProvider interface {
	IsEnabled() bool
	GetAuthSpec() istiotypes.Authentication
	GetDigdiratorName() DigdiratorName
	GetProvidedSecretName() *string
	GetIgnoredPaths() []string
	GetIssuerKey() string
	GetJwksKey() string
	GetClientIDKey() string
	GetDigdiratorClientOwnerRef(digdiratorClients DigdiratorClients) (*[]v1.OwnerReference, error)
	GetGeneratedDigdiratorSecret(digdiratorClients DigdiratorClients) (*string, error)
	HandleDigdiratorClientError(digdiratorClients DigdiratorClients) error
}
