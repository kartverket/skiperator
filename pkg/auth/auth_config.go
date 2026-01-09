package auth

import (
	"maps"
	"slices"

	"github.com/kartverket/skiperator/api/common/digdirator"
	"github.com/kartverket/skiperator/api/common/istiotypes"
)

type AuthConfigs []AuthConfig

type AuthConfig struct {
	Spec          istiotypes.RequestAuthentication
	Paths         []string
	IgnorePaths   []string
	TokenLocation string
	ProviderInfo  digdirator.DigdiratorInfo
}

func (authConfigs *AuthConfigs) GetAllPaths() []string {
	uniquePaths := map[string]struct{}{}
	if authConfigs != nil {
		for _, config := range *authConfigs {
			for _, ignoredPath := range config.IgnorePaths {
				uniquePaths[ignoredPath] = struct{}{}
			}
			for _, allowPath := range config.Paths {
				uniquePaths[allowPath] = struct{}{}
			}
		}
	}
	return slices.Collect(maps.Keys(uniquePaths))
}

func (authConfigs *AuthConfigs) GetIgnoredPaths() []string {
	ignoredPaths := map[string]string{}
	allowPaths := map[string]string{}
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
	return slices.Collect(maps.Values(ignoredPaths))
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
