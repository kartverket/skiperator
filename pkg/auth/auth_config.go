package auth

import (
	"github.com/kartverket/skiperator/api/v1alpha1/digdirator"
	"github.com/kartverket/skiperator/api/v1alpha1/istiotypes"
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
	var ignoredPaths []string
	if authConfigs != nil {
		for _, config := range *authConfigs {
			for _, ignoredPath := range config.IgnorePaths {
				if slices.Contains(ignoredPaths, ignoredPath) {
					continue
				}
				encountered := false
				for _, otherConfig := range *authConfigs {
					if slices.Contains(otherConfig.Paths, ignoredPath) {
						encountered = true
					}
				}
				if !encountered {
					if !slices.Contains(ignoredPaths, ignoredPath) {
						ignoredPaths = append(ignoredPaths, ignoredPath)
					}
				}
			}
		}
	}
	return ignoredPaths
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
