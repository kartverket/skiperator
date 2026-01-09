package istiotypes

// RequestAuthentication specifies how incoming JWTs should be validated.
//
// +kubebuilder:object:generate=true
type RequestAuthentication struct {
	// Whether to enable JWT validation.
	// If enabled, incoming JWTs will be validated against the issuer specified in the app registration and the generated audience.
	Enabled bool `json:"enabled"`

	// The name of the Kubernetes Secret containing OAuth2 credentials.
	//
	// If omitted, the associated client registration in the application manifest is used for JWT validation.
	// +kubebuilder:validation:Optional
	SecretName *string `json:"secretName,omitempty"`

	// If set to `true`, the original token will be kept for the upstream request. Defaults to `true`.
	// +kubebuilder:default=true
	ForwardJwt bool `json:"forwardJwt,omitempty"`

	// Where to find the JWT in the incoming request
	//
	// An enum value of `header` means that the JWT is present in the `Authorization` header as a `Bearer` token.
	// An enum value of `cookie` means that the JWT is present as a cookie called `BearerToken`.
	//
	// If omitted, its default value depends on the provider type:
	//   Defaults to "cookie" for providers supporting user login (e.g. IDPorten).
	//   Defaults to "header" for providers not supporting user login (e.g. Maskinporten).
	// +kubebuilder:validation:Enum=header;cookie
	// +kubebuilder:validation:Optional
	TokenLocation *string `json:"tokenLocation,omitempty"`

	// This field specifies a list of operations to copy the claim to HTTP headers on a successfully verified token.
	// The header specified in each operation in the list must be unique. Nested claims of type string/int/bool is supported as well.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:MaxItems=10
	OutputClaimToHeaders *[]ClaimToHeader `json:"outputClaimToHeaders,omitempty"`

	// Paths specifies paths that require an authenticated JWT.
	//
	// The specified paths must be a valid URI path. It has to start with '/' and cannot end with '/'.
	// The paths can also contain the wildcard operator '*', but only at the end.
	// +listType=set
	// +kubebuilder:validation:Items.Pattern=`^/[a-zA-Z0-9\-._~!$&'()+,;=:@%/]*(\*)?$`
	// +kubebuilder:validation:MaxItems=50
	// +kubebuilder:validation:Optional
	Paths *[]string `json:"paths,omitempty"`

	// IgnorePaths specifies paths that do not require an authenticated JWT.
	//
	// The specified paths must be a valid URI path. It has to start with '/' and cannot end with '/'.
	// The paths can also contain the wildcard operator '*', but only at the end.
	// +listType=set
	// +kubebuilder:validation:Items.Pattern=`^/[a-zA-Z0-9\-._~!$&'()+,;=:@%/]*(\*)?$`
	// +kubebuilder:validation:MaxItems=50
	// +kubebuilder:validation:Optional
	IgnorePaths *[]string `json:"ignorePaths,omitempty"`
}

type ClaimToHeader struct {
	// The name of the HTTP header for which the specified claim will be copied to.
	// +kubebuilder:validation:Pattern="^[a-zA-Z0-9-]+$"
	// +kubebuilder:validation:MaxLength=64
	Header string `json:"header"`

	// The claim to be copied.
	// +kubebuilder:validation:Pattern="^[a-zA-Z0-9-._]+$"
	// +kubebuilder:validation:MaxLength=128
	Claim string `json:"claim"`
}
