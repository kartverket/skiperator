package identity_provider

import (
	"github.com/kartverket/skiperator/api/v1alpha1/istiotypes"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type IdentityProviderName string

type IdentityProviderInfo struct {
	Name      IdentityProviderName
	IssuerURI string
	JwksURI   string
	ClientID  string
}

type IdentityProviderOperatorResource interface {
	GetOwnerReferences() []v1.OwnerReference
	GetSecretName() string
}

type IdentityProvider interface {
	IsRequestAuthEnabled() bool
	GetAuthSpec() *istiotypes.RequestAuth
	GetIdentityProviderName() IdentityProviderName
	GetProvidedSecretName() *string
	GetAuthRules() istiotypes.RequestAuthRules
	GetIgnoredAuthRules() istiotypes.RequestMatchers
	GetIssuerKey() string
	GetJwksKey() string
	GetClientIDKey() string
	GetTokenLocation() string
	GetAcceptedResources() []string
}
