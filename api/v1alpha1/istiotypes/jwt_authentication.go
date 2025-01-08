package istiotypes

import "istio.io/api/security/v1beta1"

// Authentication specifies how incoming JWT's should be validated.
//
// +kubebuilder:object:generate=true
type Authentication struct {
	// Whether to enable JTW validation.
	// If enabled, incoming JWT's will be validated against the issuer specified in the app registration and the generated audience.
	Enabled bool `json:"enabled"`

	// If set to true, the original token will be kept for the upstream request. Default is true.
	ForwardOriginalToken *bool `json:"forwardOriginalToken,omitempty"`

	// Where to find the JWT in the incoming request
	//
	// An enum value of HEADER means that the JWT is present in the `Authorization` as a Bearer Token.
	// An enum value of COOKIE means that the JWT is present as a cookie called `BearerToken`.
	//
	// +kubebuilder:validation:Enum=HEADER;COOKIE
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
	OutputClaimToHeaders *[]*v1beta1.ClaimToHeader `json:"outputClaimToHeaders,omitempty"`
}
