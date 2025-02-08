package jwtAuth

import (
	"context"
	"fmt"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/util"
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
		providerOps, err := GetProviderOps(providerInfo.Provider)
		if err != nil {
			return nil, fmt.Errorf("failed when retrieving provider operations for %s: %w", providerInfo.Provider, err)
		}
		secretData, err := providerOps.GetSecretData(k8sClient, ctx, application.Namespace, providerInfo.SecretName)
		if err != nil {
			return nil, fmt.Errorf("failed when retrieving secretData for %s: %w", providerInfo.Provider, err)
		}
		authConfigs = append(authConfigs, AuthConfig{
			NotPaths: providerInfo.NotPaths,
			ProviderURIs: ProviderURIs{
				Provider:  providerInfo.Provider,
				IssuerURI: string(secretData[providerOps.IssuerKey()]),
				JwksURI:   string(secretData[providerOps.JwksKey()]),
				ClientID:  string(secretData[providerOps.ClientIDKey()]),
			},
		})
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

func getIdentityProviderInfoWithAuthenticationEnabled(ctx context.Context, application skiperatorv1alpha1.Application, k8sClient client.Client) ([]IdentityProviderInfo, error) {
	var providerInfo []IdentityProviderInfo
	if util.IsIDPortenAuthenticationEnabled(application) {
		providerOps, err := GetProviderOps(ID_PORTEN)
		if err != nil {
			return nil, fmt.Errorf("failed when retrieving provider operations for %s: %w", ID_PORTEN, err)
		}
		idPortenProviderInfo, err := providerOps.GetProviderInfo(k8sClient, ctx, application)
		if err != nil {
			return nil, fmt.Errorf("failed when retrieving provider info for %s: %w", ID_PORTEN, err)
		}
		providerInfo = append(providerInfo, *idPortenProviderInfo)
	}
	if util.IsMaskinPortenAuthenticationEnabled(application) {
		providerOps, err := GetProviderOps(MASKINPORTEN)
		if err != nil {
			return nil, fmt.Errorf("failed when retrieving provider operations for %s: %w", MASKINPORTEN, err)
		}
		maskinportenProviderInfo, err := providerOps.GetProviderInfo(k8sClient, ctx, application)
		if err != nil {
			return nil, fmt.Errorf("failed when retrieving provider info for %s: %w", MASKINPORTEN, err)
		}
		providerInfo = append(providerInfo, *maskinportenProviderInfo)
	}
	return providerInfo, nil
}
