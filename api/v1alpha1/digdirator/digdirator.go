package digdirator

import (
	"github.com/kartverket/skiperator/api/v1alpha1/istiotypes"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type DigdiratorName string

type DigdiratorInfo struct {
	Name             DigdiratorName
	HostName         string
	IssuerURI        string
	JwksURI          string
	ClientID         string
	TokenURI         string
	AuthorizationURI string
	RedirectPath     string
	SignoutPath      string
}

type DigdiratorClient interface {
	GetOwnerReferences() []v1.OwnerReference
	GetSecretName() string
}

type DigdiratorProvider interface {
	IsRequestAuthEnabled() bool
	GetRequestAuthSpec() *istiotypes.RequestAuthentication
	GetDigdiratorName() DigdiratorName
	GetProvidedRequestAuthSecretName() *string
	GetRequestAuthPaths() []string
	GetRequestAuthIgnoredPaths() []string
	GetIssuerKey() string
	GetJwksKey() string
	GetClientIDKey() string
	GetTokenEndpointKey() string
	GetTokenLocation() string
}
