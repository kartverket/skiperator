package auth

import (
	"github.com/kartverket/skiperator/api/v1alpha1/digdirator"
	"github.com/kartverket/skiperator/api/v1alpha1/istiotypes"
	"golang.org/x/exp/maps"
	securityv1api "istio.io/api/security/v1"
	"slices"
)

var AcceptedHttpMethods = []string{
	"GET",
	"POST",
	"PUT",
	"PATCH",
	"DELETE",
	"HEAD",
	"OPTIONS",
	"TRACE",
	"CONNECT",
}

type AuthConfigs []AuthConfig

type AuthConfig struct {
	Spec                   istiotypes.RequestAuth
	AuthRules              istiotypes.RequestAuthRules
	IgnoreAuthRules        istiotypes.RequestMatchers
	TokenLocation          string
	AcceptedResources      []string
	IncludeInternalTraffic bool
	ProviderInfo           digdirator.DigdiratorInfo
}

func (authConfigs *AuthConfigs) IgnorePathsFromOtherAuthConfigs() {
	if authConfigs != nil {
		for index, config := range *authConfigs {
			ignoredRequestMatchers := config.FlattenOnPaths(config.IgnoreAuthRules)
			authorizedRequestMatchers := config.FlattenOnPaths(config.AuthRules.GetRequestMatchers())
			for otherIndex, otherConfig := range *authConfigs {
				if index != otherIndex {
					otherAuthorizedRequestMatchers := otherConfig.FlattenOnPaths(otherConfig.AuthRules.GetRequestMatchers())
					for otherPath, otherRequestMapper := range otherAuthorizedRequestMatchers {
						if !slices.Contains(maps.Keys(ignoredRequestMatchers), otherPath) &&
							!slices.Contains(maps.Keys(authorizedRequestMatchers), otherPath) {
							config.IgnoreAuthRules = append(config.IgnoreAuthRules, istiotypes.RequestMatcher{
								Paths:   otherRequestMapper.Operation.Paths,
								Methods: otherRequestMapper.Operation.Methods,
							})
						}
					}
				}
			}
			(*authConfigs)[index] = config
		}
	}
}

func (authConfig AuthConfig) FlattenOnPaths(requestMatchers istiotypes.RequestMatchers) map[string]*securityv1api.Rule_To {
	requestMatchersMap := make(map[string]*securityv1api.Rule_To)
	for _, requestMatcher := range requestMatchers {
		for _, path := range requestMatcher.Paths {
			if existingMatcher, exists := requestMatchersMap[path]; exists {
				// Combine methods if the path key already exists
				existingMatcher.Operation.Methods = slices.Compact(append(existingMatcher.Operation.Methods, requestMatcher.Methods...))
				requestMatchersMap[path] = existingMatcher
			} else {
				methods := requestMatcher.Methods
				if len(methods) == 0 {
					methods = AcceptedHttpMethods
				}
				requestMatchersMap[path] = &securityv1api.Rule_To{
					Operation: &securityv1api.Operation{
						Paths:   []string{path},
						Methods: methods,
					},
				}
			}
		}
	}
	return requestMatchersMap
}

func (authConfigs AuthConfigs) GetIgnoreAuthAndAuthorizedRequestMatchers() (istiotypes.RequestMatchers, istiotypes.RequestMatchers) {
	var ignoredRequestMatchers []istiotypes.RequestMatcher
	var authorizedRequestMatchers []istiotypes.RequestMatcher
	for _, authConfig := range authConfigs {
		ignoredRequestMatchers = append(ignoredRequestMatchers, authConfig.IgnoreAuthRules...)
		authorizedRequestMatchers = append(authorizedRequestMatchers, authConfig.AuthRules.GetRequestMatchers()...)
	}
	return ignoredRequestMatchers, authorizedRequestMatchers
}

func (authConfigs *AuthConfigs) GetAllPaths() []string {
	var uniquePaths map[string]struct{}
	ignoredRequestMatchers, authorizedRequestMatchers := authConfigs.GetIgnoreAuthAndAuthorizedRequestMatchers()
	for _, requestMatcher := range append(ignoredRequestMatchers, authorizedRequestMatchers...) {
		for _, path := range requestMatcher.Paths {
			uniquePaths[path] = struct{}{}
		}
	}
	return maps.Keys(uniquePaths)
}
