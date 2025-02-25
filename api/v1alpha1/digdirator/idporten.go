package digdirator

import (
	"fmt"
	"github.com/kartverket/skiperator/api/v1alpha1/istiotypes"
	"github.com/nais/digdirator/pkg/secrets"
	nais_io_v1 "github.com/nais/liberator/pkg/apis/nais.io/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const IDPortenName DigdiratorName = "idporten"

// Based off NAIS' IDPorten specification as seen here:
// https://github.com/nais/liberator/blob/c9da4cf48a52c9594afc8a4325ff49bbd359d9d2/pkg/apis/nais.io/v1/naiserator_types.go#L93C10-L93C10
//
// +kubebuilder:object:generate=true
type IDPorten struct {
	// The name of the Client as shown in Digitaliseringsdirektoratet's Samarbeidsportal
	// Meant to be a human-readable name for separating clients in the portal
	ClientName *string `json:"clientName,omitempty"`

	// Whether to enable provisioning of an ID-porten client.
	// If enabled, an ID-porten client be provisioned.
	Enabled bool `json:"enabled"`

	// AccessTokenLifetime is the lifetime in seconds for any issued access token from ID-porten.
	//
	// If unspecified, defaults to `3600` seconds (1 hour).
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=3600
	AccessTokenLifetime *int `json:"accessTokenLifetime,omitempty"`

	// ClientURI is the URL shown to the user at ID-porten when displaying a 'back' button or on errors.
	ClientURI nais_io_v1.IDPortenURI `json:"clientURI,omitempty"`

	// FrontchannelLogoutPath is a valid path for your application where ID-porten sends a request to whenever the user has
	// initiated a logout elsewhere as part of a single logout (front channel logout) process.
	//
	// +kubebuilder:validation:Pattern=`^\/.*$`
	FrontchannelLogoutPath string `json:"frontchannelLogoutPath,omitempty"`

	// IntegrationType is used to make sensible choices for your client.
	// Which type of integration you choose will provide guidance on which scopes you can use with the client.
	// A client can only have one integration type.
	//
	// NB! It is not possible to change the integration type after creation.
	//
	// +kubebuilder:validation:Enum=krr;idporten;api_klient
	IntegrationType string `json:"integrationType,omitempty" nais:"immutable"`

	// PostLogoutRedirectPath is a simpler verison of PostLogoutRedirectURIs
	// that will be appended to the ingress
	//
	// +kubebuilder:validation:Pattern=`^\/.*$`
	// +kubebuilder:validation:Optional
	PostLogoutRedirectPath string `json:"postLogoutRedirectPath,omitempty"`

	// PostLogoutRedirectURIs are valid URIs that ID-porten will allow redirecting the end-user to after a single logout
	// has been initiated and performed by the application.
	PostLogoutRedirectURIs *[]nais_io_v1.IDPortenURI `json:"postLogoutRedirectURIs,omitempty"`

	// RedirectPath is a valid path that ID-porten redirects back to after a successful authorization request.
	//
	// +kubebuilder:validation:Pattern=`^\/.*$`
	// +kubebuilder:validation:Optional
	RedirectPath string `json:"redirectPath,omitempty"`

	// Register different oauth2 Scopes on your client.
	// You will not be able to add a scope to your client that conflicts with the client's IntegrationType.
	// For example, you can not add a scope that is limited to the IntegrationType `krr` of IntegrationType `idporten`, and vice versa.
	//
	// Default for IntegrationType `krr` = ("krr:global/kontaktinformasjon.read", "krr:global/digitalpost.read")
	// Default for IntegrationType `idporten` = ("openid", "profile")
	// IntegrationType `api_klient` have no Default, checkout Digdir documentation.
	Scopes []string `json:"scopes,omitempty"`

	// SessionLifetime is the maximum lifetime in seconds for any given user's session in your application.
	// The timeout starts whenever the user is redirected from the `authorization_endpoint` at ID-porten.
	//
	// If unspecified, defaults to `7200` seconds (2 hours).
	// Note: Attempting to refresh the user's `access_token` beyond this timeout will yield an error.
	//
	// +kubebuilder:validation:Minimum=3600
	// +kubebuilder:validation:Maximum=7200
	SessionLifetime *int `json:"sessionLifetime,omitempty"`

	// AutoLogin sets up [OAuth2 authorization code flow](https://datatracker.ietf.org/doc/html/rfc6749) with ID-porten as identity provider.
	AutoLogin *istiotypes.AutoLogin `json:"autoLogin,omitempty"`

	// Authentication specifies how incoming JWT's should be validated.
	Authentication *istiotypes.RequestAuthentication `json:"requestAuthentication,omitempty"`
}

type IDPortenClient struct {
	Client nais_io_v1.IDPortenClient
}

func (i *IDPortenClient) GetSecretName() string {
	return i.Client.Spec.SecretName
}

func (i *IDPortenClient) GetOwnerReferences() []v1.OwnerReference {
	return i.Client.GetOwnerReferences()
}

func (i *IDPorten) RequestAuthEnabled() bool {
	return i != nil && i.Authentication != nil && i.Authentication.Enabled
}

func (i *IDPorten) GetRequestAuthSpec() istiotypes.RequestAuthentication {
	return *i.Authentication
}

func (i *IDPorten) GetDigdiratorName() DigdiratorName {
	return IDPortenName
}

func (i *IDPorten) GetProvidedRequestAuthSecretName() *string {
	return i.Authentication.SecretName
}

func (i *IDPorten) GetRequestAuthPaths() []string {
	var paths []string
	if i.RequestAuthEnabled() {
		if i.Authentication.Paths != nil {
			paths = append(paths, *i.Authentication.Paths...)
		}
	}
	return paths
}

func (i *IDPorten) GetRequestAuthIgnoredPaths() []string {
	var ignoredPaths []string
	if i.RequestAuthEnabled() {
		if i.Authentication.IgnorePaths != nil {
			ignoredPaths = append(ignoredPaths, *i.Authentication.IgnorePaths...)
		}
	}
	return ignoredPaths
}

func (i *IDPorten) GetIssuerKey() string {
	return secrets.IDPortenIssuerKey
}

func (i *IDPorten) GetJwksKey() string {
	return secrets.IDPortenJwksUriKey
}

func (i *IDPorten) GetClientIDKey() string {
	return secrets.IDPortenClientIDKey
}

func (i *IDPorten) GetTokenEndpointKey() string {
	return secrets.IDPortenTokenEndpointKey
}

func (i *IDPorten) AutoLoginEnabled() bool {
	return i != nil && i.AutoLogin != nil && i.AutoLogin.Enabled
}

func (i *IDPorten) GetAutoLoginSpec() istiotypes.AutoLogin {
	return *i.AutoLogin
}

func (i *IDPorten) GetProvidedAutoLoginSecretName() *string {
	return i.AutoLogin.SecretName
}

func (i *IDPorten) GetAuthorizationEndpoint() string {
	//TODO: Open PR in nais/digdirator to include authorization endpoint for ID-porten so that this can be extracted from the Secret
	return "https://login.microsoftonline.com/7f74c8a2-43ce-46b2-b0e8-b6306cba73a3/oauth2/v2.0/authorize"
}

func (i *IDPorten) GetRedirectPathKey() string {
	return secrets.IDPortenRedirectURIKey
}

func (i *IDPorten) GetSignoutPath() string {
	return i.PostLogoutRedirectPath
}

func (i *IDPorten) GetAuthScopes() ([]string, error) {
	if i.Enabled {
		if i.AutoLogin.AuthScopes != nil {
			return *i.AutoLogin.AuthScopes, nil
		}
		return i.Scopes, nil
	}
	if i.AutoLogin.AuthScopes == nil {
		return nil, fmt.Errorf("auth scopes must be specified for auto login when client registration happened outside of application manifest")
	}
	return *i.AutoLogin.AuthScopes, nil
}

func (i *IDPorten) GetAutoLoginIgnoredPaths() []string {
	if i.AutoLogin.IgnorePaths != nil {
		return *i.AutoLogin.IgnorePaths
	}
	return []string{}
}

func (i *IDPorten) GetClientSecretKey() string {
	return "IDPORTEN_CLIENT_SECRET" //TODO: This should be read by digdirator in the same way as for client id
}
