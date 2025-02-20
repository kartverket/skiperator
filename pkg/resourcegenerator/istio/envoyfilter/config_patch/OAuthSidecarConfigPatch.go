package config_patch

import "github.com/kartverket/skiperator/pkg/util"

type OAuthSidecarConfigPatchValue struct {
	Name        string      `json:"name"`
	TypedConfig TypedConfig `json:"typedConfig"`
}

type Config struct {
	TokenEndpoint         TokenEndpoint        `json:"token_endpoint"`
	AuthorizationEndpoint string               `json:"authorization_endpoint"`
	RedirectUri           string               `json:"redirect_uri"`
	RedirectPathMatcher   RedirectPathMatcher  `json:"redirect_path_matcher"`
	SignoutPath           SignoutPath          `json:"signout_path"`
	ForwardBearerToken    bool                 `json:"forward_bearer_token"`
	UseRefreshToken       bool                 `json:"use_refresh_token"`
	PassThroughMatcher    []PassThroughMatcher `json:"pass_through_matcher"`
	Credentials           Credentials          `json:"credentials"`
	AuthScopes            []string             `json:"auth_scopes"`
}

type TokenEndpoint struct {
	Cluster string `json:"cluster"`
	Uri     string `json:"uri"`
	Timeout string `json:"timeout"`
}

type RedirectPathMatcher struct {
	Path Path `json:"path"`
}

type SignoutPath struct {
	Path Path `json:"path"`
}

type Path struct {
	Exact string `json:"exact"`
}

type PassThroughMatcher struct {
	Name        string      `json:"name"`
	StringMatch StringMatch `json:"string_match"`
}

type StringMatch struct {
	Prefix string `json:"prefix"`
}

type Credentials struct {
	ClientId    string            `json:"client_id"`
	TokenSecret SecretCredentials `json:"token_secret"`
	HmacSecret  SecretCredentials `json:"hmac_secret"`
}

type SecretCredentials struct {
	Name      string    `json:"name"`
	SdsConfig SdsConfig `json:"sds_config"`
}

type SdsConfig struct {
	PathConfigSource PathConfigSource `json:"path_config_source"`
}

type PathConfigSource struct {
	Path             string           `json:"path"`
	WatchedDirectory WatchedDirectory `json:"watched_directory"`
}

type WatchedDirectory struct {
	Path string `json:"path"`
}

func GetOAuthSidecarConfigPatchValue(tokenEndpoint string, authorizationEndpoint string, redirectPath string, signoutPath string, ignorePaths []string, clientId string, authScopes []string) OAuthSidecarConfigPatchValue {
	passThroughMatchers := []PassThroughMatcher{
		{
			Name: "authorization",
			StringMatch: StringMatch{
				Prefix: "Bearer ",
			},
		},
	}
	for _, ignorePath := range ignorePaths {
		passThroughMatchers = append(passThroughMatchers, PassThroughMatcher{
			Name: ":path",
			StringMatch: StringMatch{
				Prefix: ignorePath,
			},
		})
	}
	return OAuthSidecarConfigPatchValue{
		Name: "envoy.filters.http.oauth2",
		TypedConfig: TypedConfig{
			Type: "type.googleapis.com/envoy.extensions.filters.http.oauth2.v3.OAuth2",
			Config: &Config{
				TokenEndpoint: TokenEndpoint{
					Cluster: "oauth",
					Uri:     tokenEndpoint,
					Timeout: "5s",
				},
				AuthorizationEndpoint: authorizationEndpoint,
				RedirectUri:           "https://%REQ(:authority)%" + redirectPath,
				RedirectPathMatcher: RedirectPathMatcher{
					Path: Path{
						Exact: redirectPath,
					},
				},
				SignoutPath: SignoutPath{
					Path: Path{
						Exact: signoutPath,
					},
				},
				ForwardBearerToken: true,
				UseRefreshToken:    true,
				PassThroughMatcher: passThroughMatchers,
				Credentials: Credentials{
					ClientId: clientId,
					TokenSecret: SecretCredentials{
						Name: "envoy-client",
						SdsConfig: SdsConfig{
							PathConfigSource: PathConfigSource{
								Path: util.IstioTokenSecretSource,
								WatchedDirectory: WatchedDirectory{
									Path: util.IstioCredentialsDirectory,
								},
							},
						},
					},
					HmacSecret: SecretCredentials{
						Name: "hmac",
						SdsConfig: SdsConfig{
							PathConfigSource: PathConfigSource{
								Path: util.IstioHmacSecretSource,
								WatchedDirectory: WatchedDirectory{
									Path: util.IstioCredentialsDirectory,
								},
							},
						},
					},
				},
				AuthScopes: authScopes,
			},
		},
	}
}
