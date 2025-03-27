package digdirator

import (
	"github.com/kartverket/skiperator/v3/api/v1alpha1/istiotypes"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type DigdiratorName string

type DigdiratorInfo struct {
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
	IsRequestAuthEnabled() bool
	GetAuthSpec() *istiotypes.RequestAuthentication
	GetDigdiratorName() DigdiratorName
	GetProvidedSecretName() *string
	GetPaths() []string
	GetIgnoredPaths() []string
	GetIssuerKey() string
	GetJwksKey() string
	GetClientIDKey() string
	GetTokenLocation() string
}
