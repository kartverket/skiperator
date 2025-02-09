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

type IDPortenOps struct{}

func (i *IDPortenOps) IsEnabled(application skiperatorv1alpha1.Application) bool {
	return util.IsIDPortenAuthenticationEnabled(application)
}

func (i *IDPortenOps) Provider() IdentityProvider {
	return ID_PORTEN
}

func (i *IDPortenOps) GetSecret(k8sClient client.Client, ctx context.Context, application skiperatorv1alpha1.Application) (*v1.Secret, error) {
	if i.IsEnabled(application) {
		namespacedName := types.NamespacedName{
			Namespace: application.Namespace,
		}
		if application.Spec.IDPorten.Authentication.SecretName != nil {
			namespacedName.Name = *application.Spec.IDPorten.Authentication.SecretName
			secret, err := util.GetSecret(k8sClient, ctx, namespacedName)
			if err != nil {
				return nil, fmt.Errorf("failed to retrieve IDPorten secret: '%s': %w", namespacedName, err)
			}
			return &secret, nil
		}
		namespacedName.Name = application.Name
		idPortenClient, err := util.GetIdPortenClient(k8sClient, ctx, namespacedName)
		if err != nil {
			return nil, fmt.Errorf("failed to get IDPortenClient: %s", namespacedName.String())
		}
		if idPortenClient == nil {
			return nil, fmt.Errorf("IDPortenClient: '%s' not found", namespacedName.String())
		}
		for _, ownerReference := range idPortenClient.OwnerReferences {
			if ownerReference.UID == application.UID {
				secretName := types.NamespacedName{
					Namespace: application.Namespace,
					Name:      idPortenClient.Spec.SecretName,
				}
				secret, err := util.GetSecret(k8sClient, ctx, secretName)
				if err != nil {
					return nil, fmt.Errorf("failed to retrieve IDPorten secret: '%s': %w", secretName.String(), err)
				}
				return &secret, nil
			}
		}
		return nil, fmt.Errorf("no IDPortenClient with ownerRef to '%s' found", namespacedName.String())
	} else {
		return nil, fmt.Errorf("IDPorten authentication not enabled for application: (%s, %s)", application.Name, application.Namespace)
	}
}

func (i *IDPortenOps) GetIgnoredPaths(application skiperatorv1alpha1.Application) []string {
	var ignoredPaths []string
	if i.IsEnabled(application) {
		if application.Spec.IDPorten.Authentication.IgnorePaths != nil {
			ignoredPaths = append(ignoredPaths, *application.Spec.IDPorten.Authentication.IgnorePaths...)
		}
	}
	return ignoredPaths
}

func (i *IDPortenOps) GetAuthConfig(k8sClient client.Client, ctx context.Context, application skiperatorv1alpha1.Application) (*AuthConfig, error) {
	if i.IsEnabled(application) {
		secret, err := i.GetSecret(k8sClient, ctx, application)
		if err != nil {
			return nil, fmt.Errorf("failed to get IDPorten secret: %w", err)
		}
		return &AuthConfig{
			NotPaths: i.GetIgnoredPaths(application),
			ProviderURIs: ProviderURIs{
				Provider:  i.Provider(),
				IssuerURI: string(secret.Data[secrets.IDPortenIssuerKey]),
				JwksURI:   string(secret.Data[secrets.IDPortenJwksUriKey]),
				ClientID:  string(secret.Data[secrets.IDPortenClientIDKey]),
			},
		}, nil
	}
	return nil, fmt.Errorf("IDPorten authentication not enabled for application: (%s, %s)", application.Name, application.Namespace)
}
