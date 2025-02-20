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

type DigdiratorClient interface {
	GetOwnerReferences() []v1.OwnerReference
	GetSecretName() string
}

type DigdiratorProvider interface {
	IsEnabled() bool
	GetAuthSpec() istiotypes.RequestAuthentication
	GetDigdiratorName() DigdiratorName
	GetProvidedSecretName() *string
	GetPaths() []string
	GetIgnoredPaths() []string
	GetIssuerKey() string
	GetJwksKey() string
	GetClientIDKey() string
}
