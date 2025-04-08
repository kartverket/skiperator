package idprovider

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

type IdentityProviderOperatorCRD interface {
	GetOwnerReferences() []v1.OwnerReference
	GetSecretName() string
}

type IdentityProvider interface {
	IsRequestAuthEnabled() bool
	GetAuthSpec() *istiotypes.RequestAuthentication
	GetIdentityProviderName() IdentityProviderName
	GetProvidedSecretName() *string
	GetPaths() []string
	GetIgnoredPaths() []string
	GetIssuerKey() string
	GetJwksKey() string
	GetClientIDKey() string
	GetTokenLocation() string
}
