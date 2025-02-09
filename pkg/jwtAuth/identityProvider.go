package jwtAuth

import (
	"context"
	"github.com/kartverket/skiperator/api/v1alpha1"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type IdentityProvider string

var (
	MASKINPORTEN IdentityProvider = "MASKINPORTEN"
	ID_PORTEN    IdentityProvider = "ID_PORTEN"
)

type ProviderURIs struct {
	Provider  IdentityProvider
	IssuerURI string
	JwksURI   string
	ClientID  string
}

type ProviderOps interface {
	IsEnabled(application v1alpha1.Application) bool
	Provider() IdentityProvider
	GetSecret(k8sClient client.Client, ctx context.Context, application v1alpha1.Application) (*v1.Secret, error)
	GetIgnoredPaths(application v1alpha1.Application) []string
	GetAuthConfig(k8sClient client.Client, ctx context.Context, application v1alpha1.Application) (*AuthConfig, error)
}

func GetProviderOps() []ProviderOps {
	return []ProviderOps{
		&IDPortenOps{},
		&MaskinportenOps{},
	}
}
