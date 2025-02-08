package jwtAuth

import (
	"context"
	"fmt"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/util"
	"github.com/nais/digdirator/pkg/secrets"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type AuthConfigs []AuthConfig

type AuthConfig struct {
	NotPaths     *[]string
	ProviderURIs ProviderURIs
}

type ProviderURIs struct {
	Provider  IdentityProvider
	IssuerURI string
	JwksURI   string
	ClientID  string
}

var (
	MASKINPORTEN IdentityProvider = "MASKINPORTEN"
	ID_PORTEN    IdentityProvider = "ID_PORTEN"
)

type IdentityProvider string

type IdentityProviderInfo struct {
	Provider   IdentityProvider
	SecretName string
	NotPaths   *[]string
}

func GetAuthConfigsForApplication(k8sClient client.Client, ctx context.Context, application *skiperatorv1alpha1.Application) (*AuthConfigs, error) {
	identityProviderInfo, err := getIdentityProviderInfoWithAuthenticationEnabled(ctx, application, k8sClient)
	if err != nil {
		return nil, fmt.Errorf("failed when getting identity provider info: %w", err)
	}
	var authConfigs AuthConfigs
	for _, providerInfo := range identityProviderInfo {
		switch providerInfo.Provider {
		case ID_PORTEN:
			secretData, err := util.GetSecretData(k8sClient, ctx, types.NamespacedName{
				Namespace: application.Namespace,
				Name:      providerInfo.SecretName,
			}, []string{secrets.IDPortenIssuerKey, secrets.IDPortenJwksUriKey, secrets.IDPortenClientIDKey})
			if err != nil {
				return nil, fmt.Errorf("failed when retrieving idporten secretData: %w", err)
			}
			authConfigs = append(authConfigs, AuthConfig{
				NotPaths: providerInfo.NotPaths,
				ProviderURIs: ProviderURIs{
					Provider:  ID_PORTEN,
					IssuerURI: string(secretData[secrets.IDPortenIssuerKey]),
					JwksURI:   string(secretData[secrets.IDPortenJwksUriKey]),
					ClientID:  string(secretData[secrets.IDPortenClientIDKey]),
				},
			})
		case MASKINPORTEN:
			secretData, err := util.GetSecretData(k8sClient, ctx, types.NamespacedName{
				Namespace: application.Namespace,
				Name:      providerInfo.SecretName,
			}, []string{secrets.MaskinportenIssuerKey, secrets.MaskinportenJwksUriKey, secrets.MaskinportenClientIDKey})
			if err != nil {
				return nil, fmt.Errorf("failed when retrieving maskinporten secretData: %w", err)
			}
			authConfigs = append(authConfigs, AuthConfig{
				NotPaths: providerInfo.NotPaths,
				ProviderURIs: ProviderURIs{
					Provider:  MASKINPORTEN,
					IssuerURI: string(secretData[secrets.MaskinportenIssuerKey]),
					JwksURI:   string(secretData[secrets.MaskinportenJwksUriKey]),
					ClientID:  string(secretData[secrets.MaskinportenClientIDKey]),
				},
			})
		default:
			return nil, fmt.Errorf("unknown provider: %s", providerInfo.Provider)
		}
	}
	if len(authConfigs) > 0 {
		return &authConfigs, nil
	} else {
		return nil, nil
	}
}

func (authConfigs *AuthConfigs) GetAllowedPaths(authorizationSettings *skiperatorv1alpha1.AuthorizationSettings) []string {
	allowPaths := []string{}
	if authorizationSettings != nil {
		if authorizationSettings.AllowList != nil {
			if len(authorizationSettings.AllowList) > 0 {
				allowPaths = authorizationSettings.AllowList
			}
		}
	}
	if authConfigs != nil {
		for _, config := range *authConfigs {
			if config.NotPaths != nil {
				allowPaths = append(allowPaths, *config.NotPaths...)
			}
		}
	}
	return allowPaths
}

func getIdentityProviderInfoWithAuthenticationEnabled(ctx context.Context, application *skiperatorv1alpha1.Application, k8sClient client.Client) ([]IdentityProviderInfo, error) {
	var providerInfo []IdentityProviderInfo
	if util.IsIDPortenAuthenticationEnabled(application) {
		var secretName *string
		var err error
		if application.Spec.IDPorten.Authentication.SecretName != nil {
			// If secret name is provided, use it regardless of whether IDPorten is enabled
			secretName = application.Spec.IDPorten.Authentication.SecretName
		} else if application.Spec.IDPorten.Enabled {
			// If IDPorten is enabled but no secretName provided, retrieve the generated secret from IDPortenClient
			secretName, err = getSecretNameForIdentityProvider(k8sClient, ctx,
				types.NamespacedName{
					Namespace: application.Namespace,
					Name:      application.Name,
				},
				ID_PORTEN,
				application.UID)
		} else {
			// If IDPorten is not enabled and no secretName provided, return error
			return nil, fmt.Errorf("JWT authentication requires either IDPorten to be enabled or a secretName to be provided")
		}
		if err != nil {
			err := fmt.Errorf("failed to get secret name for IDPortenClient: %w", err)
			return nil, err
		}

		var notPaths *[]string
		if application.Spec.IDPorten.Authentication.IgnorePaths != nil {
			notPaths = application.Spec.IDPorten.Authentication.IgnorePaths
		} else {
			notPaths = nil
		}
		providerInfo = append(providerInfo, IdentityProviderInfo{
			Provider:   ID_PORTEN,
			SecretName: *secretName,
			NotPaths:   notPaths,
		})
	}
	if util.IsMaskinPortenAuthenticationEnabled(application) {
		var secretName *string
		var err error
		if application.Spec.Maskinporten.Authentication.SecretName != nil {
			// If secret name is provided, use it regardless of whether Maskinporten is enabled
			secretName = application.Spec.Maskinporten.Authentication.SecretName
		} else if application.Spec.Maskinporten.Enabled {
			// If Maskinporten is enabled but no secretName provided, retrieve the generated secret from MaksinPortenClient
			secretName, err = getSecretNameForIdentityProvider(k8sClient, ctx,
				types.NamespacedName{
					Namespace: application.Namespace,
					Name:      application.Name,
				},
				MASKINPORTEN,
				application.UID)
		} else {
			// If Maskinporten is not enabled and no secretName provided, return error
			return nil, fmt.Errorf("JWT authentication requires either Maskinporten to be enabled or a secretName to be provided")
		}
		if err != nil {
			err := fmt.Errorf("failed to get secret name for MaskinPortenClient: %w", err)
			return nil, err
		}

		var notPaths *[]string
		if application.Spec.Maskinporten.Authentication.IgnorePaths != nil {
			notPaths = application.Spec.Maskinporten.Authentication.IgnorePaths
		} else {
			notPaths = nil
		}
		providerInfo = append(providerInfo, IdentityProviderInfo{
			Provider:   MASKINPORTEN,
			SecretName: *secretName,
			NotPaths:   notPaths,
		})
	}
	return providerInfo, nil
}

func getSecretNameForIdentityProvider(k8sClient client.Client, ctx context.Context, namespacedName types.NamespacedName, provider IdentityProvider, applicationUID types.UID) (*string, error) {
	switch provider {
	case ID_PORTEN:
		idPortenClient, err := util.GetIdPortenClient(k8sClient, ctx, namespacedName)
		if err != nil {
			err := fmt.Errorf("failed to get IDPortenClient: %s", namespacedName.String())
			return nil, err
		}
		if idPortenClient == nil {
			err := fmt.Errorf("IDPortenClient: '%s' not found", namespacedName.String())
			return nil, err
		}
		for _, ownerReference := range idPortenClient.OwnerReferences {
			if ownerReference.UID == applicationUID {
				return &idPortenClient.Spec.SecretName, nil
			}
		}
		err = fmt.Errorf("no IDPortenClient with ownerRef to '%s' found", namespacedName.String())
		return nil, err

	case MASKINPORTEN:
		maskinPortenClient, err := util.GetMaskinPortenlient(k8sClient, ctx, namespacedName)
		if err != nil {
			err := fmt.Errorf("failed to get MaskinPortenClient: %s", namespacedName.String())
			return nil, err
		}
		if maskinPortenClient == nil {
			err := fmt.Errorf("IDPortenClient: '%s' not found", namespacedName.String())
			return nil, err
		}
		for _, ownerReference := range maskinPortenClient.OwnerReferences {
			if ownerReference.UID == applicationUID {
				return &maskinPortenClient.Spec.SecretName, nil
			}
		}
		err = fmt.Errorf("no MaskinPortenClient with ownerRef to (%s) found", namespacedName.String())
		return nil, err

	default:
		return nil, fmt.Errorf("provider: %s not supported", provider)
	}
}
