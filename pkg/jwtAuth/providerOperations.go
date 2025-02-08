package jwtAuth

import (
	"context"
	"fmt"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/jwtAuth/idporten"
	"github.com/kartverket/skiperator/pkg/jwtAuth/maskinporten"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ProviderOps interface {
	Provider() IdentityProvider
	IssuerKey() string
	JwksKey() string
	ClientIDKey() string
	GetSecretName(k8sClient client.Client, ctx context.Context, application skiperatorv1alpha1.Application) (*string, error)
	GetSecretData(k8sClient client.Client, ctx context.Context, namespace, secretName string) (map[string][]byte, error)
	GetProviderInfo(k8sClient client.Client, ctx context.Context, application skiperatorv1alpha1.Application) (*IdentityProviderInfo, error)
}

func GetProviderOps(provider IdentityProvider) (ProviderOps, error) {
	switch provider {
	case ID_PORTEN:
		return &idporten.IDPortenOps{}, nil
	case MASKINPORTEN:
		return &maskinporten.MaskinportenOps{}, nil
	default:
		return nil, fmt.Errorf("unknown provider: %s", provider)
	}
}
