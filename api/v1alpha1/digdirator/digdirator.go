package digdirator

import (
	"github.com/kartverket/skiperator/api/v1alpha1/istiotypes"
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
	GetAuthSpec() *istiotypes.RequestAuth
	GetDigdiratorName() DigdiratorName
	GetProvidedSecretName() *string
	GetAuthRules() istiotypes.RequestAuthRules
	GetIgnoredAuthRules() istiotypes.RequestMatchers
	GetIssuerKey() string
	GetJwksKey() string
	GetClientIDKey() string
	GetTokenLocation() string
	GetAcceptedResources() []string
}
