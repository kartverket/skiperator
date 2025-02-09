package jwtAuth

import (
	"context"
	"fmt"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type AuthConfigs []AuthConfig

type AuthConfig struct {
	NotPaths     []string
	ProviderURIs ProviderURIs
}

func GetAuthConfigsForApplication(k8sClient client.Client, ctx context.Context, application *skiperatorv1alpha1.Application) (*AuthConfigs, error) {
	if application == nil {
		return nil, fmt.Errorf("cannot retrieve AuthConfigs for nil application")
	}
	var authConfigs AuthConfigs

	for _, providerOps := range GetProviderOps() {
		if providerOps.IsEnabled(*application) {
			authConfig, err := providerOps.GetAuthConfig(k8sClient, ctx, *application)
			if err != nil {
				return nil, fmt.Errorf("could not get auth config for provider '%s': %w", providerOps.Provider(), err)
			}
			authConfigs = append(authConfigs, *authConfig)
		}
	}

	if len(authConfigs) > 0 {
		return &authConfigs, nil
	} else {
		return nil, nil
	}
}

func (authConfigs *AuthConfigs) GetAllowedPaths(authorizationSettings *skiperatorv1alpha1.AuthorizationSettings) []string {
	var allowPaths []string
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
				allowPaths = append(allowPaths, config.NotPaths...)
			}
		}
	}
	return allowPaths
}
