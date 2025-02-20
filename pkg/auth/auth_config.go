package auth

import (
	"github.com/kartverket/skiperator/api/v1alpha1/digdirator"
	"github.com/kartverket/skiperator/api/v1alpha1/istiotypes"
	"golang.org/x/exp/maps"
	"slices"
)

type AuthConfigs []AuthConfig

type AuthConfig struct {
	Spec         istiotypes.Authentication
	Paths        []string
	IgnorePaths  []string
	ProviderURIs digdirator.DigdiratorURIs
}

func (authConfigs *AuthConfigs) GetIgnoredPaths() []string {
	var ignoredPaths map[string]string
	var allowPaths map[string]string
	if authConfigs != nil {
		for _, config := range *authConfigs {
			for _, ignoredPath := range config.IgnorePaths {
				ignoredPaths[ignoredPath] = ignoredPath
			}
			for _, allowPath := range config.Paths {
				allowPaths[allowPath] = allowPath
			}
		}

		for _, path := range allowPaths {
			if _, ok := ignoredPaths[path]; ok {
				delete(ignoredPaths, path)
			}

		}
	}
	return maps.Values(ignoredPaths)
}

func (authConfigs *AuthConfigs) IgnorePathsFromOtherAuthConfigs() {
	if authConfigs != nil {
		for index, config := range *authConfigs {
			for otherIndex, otherConfig := range *authConfigs {
				if index != otherIndex {
					for _, otherPath := range otherConfig.Paths {
						if !slices.Contains(config.IgnorePaths, otherPath) && !slices.Contains(config.Paths, otherPath) {
							config.IgnorePaths = append(config.IgnorePaths, otherPath)
						}
					}
				}
			}
			(*authConfigs)[index] = config
		}
	}
}
