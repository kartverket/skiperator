package config_patch

import "strings"

func GetOAuthSidecarConfigPatchValue(tokenEndpoint string, authorizationEndpoint string, redirectPath string, signoutPath string, ignorePaths []string, clientId string, authScopes []string) map[string]interface{} {
	passThroughMatchers := []interface{}{
		map[string]interface{}{
			"name": "authorization",
			"string_match": map[string]interface{}{
				"prefix": "Bearer ",
			},
		},
	}

	// Add ignore paths to pass-through matchers
	for _, ignorePath := range ignorePaths {
		passThroughMatchers = addPassThroughMatcher(passThroughMatchers, ignorePath)
	}

	// Convert authScopes []string into []interface{}
	authScopesInterface := make([]interface{}, len(authScopes))
	for i, v := range authScopes {
		authScopesInterface[i] = v
	}

	return map[string]interface{}{
		"name": "envoy.filters.http.oauth2",
		"typed_config": map[string]interface{}{
			"@type": "type.googleapis.com/envoy.extensions.filters.http.oauth2.v3.OAuth2",
			"config": map[string]interface{}{
				"token_endpoint": map[string]interface{}{
					"cluster": "oauth",
					"uri":     tokenEndpoint,
					"timeout": "5s",
				},
				"authorization_endpoint": authorizationEndpoint,
				"redirect_uri":           "https://%REQ(:authority)%" + redirectPath,
				"redirect_path_matcher": map[string]interface{}{
					"path": map[string]interface{}{
						"exact": redirectPath,
					},
				},
				"signout_path": map[string]interface{}{
					"path": map[string]interface{}{
						"exact": signoutPath,
					},
				},
				"forward_bearer_token": true,
				"use_refresh_token":    true,
				"pass_through_matcher": passThroughMatchers,
				"credentials": map[string]interface{}{
					"client_id": clientId,
					"token_secret": map[string]interface{}{
						"name": "token",
						"sds_config": map[string]interface{}{
							"path_config_source": map[string]interface{}{
								"path": "/etc/istio/config/token-secret.yaml",
								"watched_directory": map[string]interface{}{
									"path": "/etc/istio/config",
								},
							},
						},
					},
					"hmac_secret": map[string]interface{}{
						"name": "hmac",
						"sds_config": map[string]interface{}{
							"path_config_source": map[string]interface{}{
								"path": "/etc/istio/config/hmac-secret.yaml",
								"watched_directory": map[string]interface{}{
									"path": "/etc/istio/config",
								},
							},
						},
					},
				},
				"auth_scopes": authScopesInterface,
			},
		},
	}
}

func addPassThroughMatcher(passThroughMatchers []interface{}, ignorePath string) []interface{} {
	stringMatch := map[string]interface{}{}
	if strings.HasSuffix(ignorePath, "*") {
		stringMatch["prefix"] = strings.TrimSuffix(ignorePath, "*")
	} else {
		stringMatch["exact"] = ignorePath
	}
	return append(passThroughMatchers, map[string]interface{}{
		"name":         ":path",
		"string_match": stringMatch,
	})
}
