package istiotypes

// AutoLogin sets up [OAuth2 authorization code flow](https://datatracker.ietf.org/doc/html/rfc6749) for the application.
//
// +kubebuilder:object:generate=true
type AutoLogin struct {
	// Whether to enable auto login.
	// If enabled, requests that do not include a JWT in the 'BearerToken' cookie
	// will be redirected to the user login page.
	Enabled bool `json:"enabled"`

	// The name of the kubernetes Secret containing OAuth2-credentials.
	//
	// If omitted, the associated client registration in the application manifest is used for user login.
	SecretName *string `json:"secretName,omitempty"`

	// IgnorePaths specifies the routes that do not require a JWT and will not redirect to the user login page if absent.
	//
	// The specified paths must start with '/'.
	// +listType=set
	// +kubebuilder:validation:Items.Pattern="^/"
	// +kubebuilder:validation:MaxItems=50
	IgnorePaths *[]string `json:"ignorePaths,omitempty"`

	// AuthScopes specifies which scopes that should be used when the application authorizes towards the identity provider.
	//
	// The specified scopes must be a sub-set of the scopes granted for the app-registration.
	// Required when the associated client registration is not a part of the application manifest.
	// If omitted, the scopes from the associated client registration in the application manifest is used.
	// +listType=set
	// +kubebuilder:validation:MaxItems=10
	AuthScopes *[]string `json:"authScopes,omitempty"`
}
