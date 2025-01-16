package istiotypes

// Authentication specifies how incoming JWT's should be validated.
//
// +kubebuilder:object:generate=true
type Authentication struct {
	// Whether to enable JWT validation.
	// If enabled, incoming JWT's will be validated against the issuer specified in the app registration and the generated audience.
	Enabled bool `json:"enabled"`

	// The name of the kubernetes Secret containing OAuth2-credentials.
	//
	// If omitted, the associated client registration in the application manifest is used for JWT validation.
	SecretName *string `json:"secretName,omitempty"`

	// If set to true, the original token will be kept for the upstream request. Default is true.
	ForwardOriginalToken *bool `json:"forwardOriginalToken,omitempty"`

	// Where to find the JWT in the incoming request
	//
	// An enum value of `header` means that the JWT is present in the `Authorization` as a Bearer Token.
	// An enum value of `cookie` means that the JWT is present as a cookie called `BearerToken`.
	//
	// +kubebuilder:validation:Enum=header;cookie
	TokenLocation *string `json:"tokenLocation,omitempty"`

	// This field specifies a list of operations to copy the claim to HTTP headers on a successfully verified token.
	// The header specified in each operation in the list must be unique. Nested claims of type string/int/bool is supported as well.
	// ```
	//
	//	outputClaimToHeaders:
	//	- header: x-my-company-jwt-group
	//	  claim: my-group
	//	- header: x-test-environment-flag
	//	  claim: test-flag
	//	- header: x-jwt-claim-group
	//	  claim: nested.key.group
	//
	// ```
	OutputClaimToHeaders *[]ClaimToHeader `json:"outputClaimToHeaders,omitempty"`

	// IgnorePaths specifies paths that do not require an authenticated JWT.
	//
	// The specified paths must start with '/'.
	// +listType=set
	// +kubebuilder:validation:Items.Pattern="^/"
	IgnorePaths *[]string `json:"ignorePaths,omitempty"`
}

type ClaimToHeader struct {
	// The name of the HTTP-header for which the specified claim will be copied to.
	Header string `json:"header"`

	// The claim to be copied.
	Claim string `json:"claim"`
}
