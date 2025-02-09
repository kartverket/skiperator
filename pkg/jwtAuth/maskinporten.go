package jwtAuth

import (
	"context"
	"fmt"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/util"
	"github.com/nais/digdirator/pkg/secrets"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type MaskinportenOps struct{}

func (i *MaskinportenOps) IsEnabled(application skiperatorv1alpha1.Application) bool {
	return util.IsMaskinPortenAuthenticationEnabled(application)
}

func (i *MaskinportenOps) Provider() IdentityProvider {
	return MASKINPORTEN
}

func (i *MaskinportenOps) GetSecret(k8sClient client.Client, ctx context.Context, application skiperatorv1alpha1.Application) (*v1.Secret, error) {
	if i.IsEnabled(application) {
		namespacedName := types.NamespacedName{
			Namespace: application.Namespace,
		}
		if application.Spec.Maskinporten.Authentication.SecretName != nil {
			namespacedName.Name = *application.Spec.Maskinporten.Authentication.SecretName
			secret, err := util.GetSecret(k8sClient, ctx, namespacedName)
			if err != nil {
				return nil, fmt.Errorf("failed to retrieve Maskinporten secret: '%s': %w", namespacedName, err)
			}
			return &secret, nil
		}
		namespacedName.Name = application.Name
		maskinportenClient, err := util.GetMaskinPortenlient(k8sClient, ctx, namespacedName)
		if err != nil {
			return nil, fmt.Errorf("failed to get MaskinportenClient: %s", namespacedName.String())
		}
		if maskinportenClient == nil {
			return nil, fmt.Errorf("MaskinportenClient: '%s' not found", namespacedName.String())
		}
		for _, ownerReference := range maskinportenClient.OwnerReferences {
			if ownerReference.UID == application.UID {
				secretName := types.NamespacedName{
					Namespace: application.Namespace,
					Name:      maskinportenClient.Spec.SecretName,
				}
				secret, err := util.GetSecret(k8sClient, ctx, secretName)
				if err != nil {
					return nil, fmt.Errorf("failed to retrieve Maskinporten secret: '%s': %w", secretName.String(), err)
				}
				return &secret, nil
			}
		}
		return nil, fmt.Errorf("no MaskinportenClient with ownerRef to '%s' found", namespacedName.String())
	} else {
		return nil, fmt.Errorf("maskinporten authentication not enabled for application: (%s, %s)", application.Name, application.Namespace)
	}
}

func (i *MaskinportenOps) GetIgnoredPaths(application skiperatorv1alpha1.Application) []string {
	var ignoredPaths []string
	if i.IsEnabled(application) {
		if application.Spec.Maskinporten.Authentication.IgnorePaths != nil {
			ignoredPaths = append(ignoredPaths, *application.Spec.Maskinporten.Authentication.IgnorePaths...)
		}
	}
	return ignoredPaths
}

func (i *MaskinportenOps) GetAuthConfig(k8sClient client.Client, ctx context.Context, application skiperatorv1alpha1.Application) (*AuthConfig, error) {
	if i.IsEnabled(application) {
		secret, err := i.GetSecret(k8sClient, ctx, application)
		if err != nil {
			return nil, fmt.Errorf("failed to get Maskinporten secret: %w", err)
		}
		return &AuthConfig{
			NotPaths: i.GetIgnoredPaths(application),
			ProviderURIs: ProviderURIs{
				Provider:  i.Provider(),
				IssuerURI: string(secret.Data[secrets.MaskinportenIssuerKey]),
				JwksURI:   string(secret.Data[secrets.MaskinportenJwksUriKey]),
				ClientID:  string(secret.Data[secrets.MaskinportenClientIDKey]),
			},
		}, nil
	}
	return nil, fmt.Errorf("maskinporten authentication not enabled for application: (%s, %s)", application.Name, application.Namespace)
}
