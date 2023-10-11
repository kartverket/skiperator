package digdirator

import nais_io_v1 "github.com/nais/liberator/pkg/apis/nais.io/v1"

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
}
