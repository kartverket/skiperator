package identity_provider

import (
	"fmt"
	"github.com/kartverket/skiperator/api/v1alpha1/istiotypes"
	nais_io_v1 "github.com/nais/liberator/pkg/apis/nais.io/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const EntraIDName IdentityProviderName = "entraid"

// Based off NAIS' AzureAdApplication specification as seen here:
// https://github.com/nais/liberator/blob/2399ce86b625e56a633af9284ddda08bc3933f7b/pkg/apis/nais.io/v1/azureadapplication_types.go
//
// +kubebuilder:object:generate=true
type EntraID struct {
	// Whether to enable provisioning of an AzureAdApplication resource.
	// If enabled, an AzureAdApplication resource will be provisioned.
	Enabled bool `json:"enabled"`

	// AllowAllUsers denotes whether all users within Azurerator's configured tenant should be allowed to access this AzureAdApplication. Defaults to false.
	//
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=false
	AllowAllUsers bool `json:"allowAllUsers,omitempty"`

	// Claims defines additional configuration of the emitted claims in tokens returned to the Azure AD application.
	//
	// +kubebuilder:validation:Optional
	Claims *EntraIDClaims `json:"claims,omitempty"`

	// LogoutUrl is the URL where Entra ID sends a request to have the application clear the user's session data.
	// This is required if single sign-out should work correctly. Must start with 'https'.
	// Defaults to the first configured ingress with https and "/oauth2/logout" as the path if not set.
	//
	// +kubebuilder:validation:Optional
	LogoutUrl *string `json:"logoutUrl,omitempty"`

	PreAuthorizedApplications []nais_io_v1.AccessPolicyInboundRule `json:"preAuthorizedApplications,omitempty"`

	// ReplyUrls defines URLs which the applications accept as reply URLs after authenticating.
	// Defaults to all configured ingresses with "/oauth2/callback" as the path if not set.
	//
	// +kubebuilder:validation:Optional
	ReplyUrls []nais_io_v1.AzureAdReplyUrl `json:"replyUrls,omitempty"`

	// SecretName is the name of the resulting Secret resource to be created and injected in the deployment.
	// Defaults to a pseudo-random name to avoid collisions.
	//
	// +kubebuilder:validation:Optional
	SecretName *string `json:"secretName,omitempty"`

	// SecretKeyPrefix is an optional user-defined prefix applied to the keys in the secret output, replacing the default prefix which is "AZURE".
	//
	// +kubebuilder:validation:Optional
	SecretKeyPrefix *string `json:"secretKeyPrefix,omitempty"`

	// SecretProtected protects the secret's credentials from being revoked by the janitor even when not in use.
	//
	// +kubebuilder:validation:Optional
	SecretProtected *bool `json:"secretProtected,omitempty"`

	// SinglePageApplication denotes whether or not the Entra ID application should be registered as a single-page-application for usage in client-side applications without access to secrets.
	//
	// +kubebuilder:validation:Optional
	SinglePageApplication *bool `json:"singlePageApplication,omitempty"`

	// RequestAuthentication specifies how incoming JWTs should be validated.
	RequestAuthentication *istiotypes.RequestAuthentication `json:"requestAuthentication,omitempty"`
}

// EntraIDClaims defines additional configuration of the emitted claims in tokens returned to the Entra ID application.
// +kubebuilder:object:generate=true
type EntraIDClaims struct {
	// Groups is a list of Entra ID group IDs to be emitted in the `groups` claim in tokens issued by Entra ID.
	// This also assigns groups to the application for access control. Only direct members of the groups are granted access.
	Groups []nais_io_v1.AzureAdGroup `json:"groups,omitempty"`
}

type AzureAdApplication struct {
	Resource nais_io_v1.AzureAdApplication
}

func (i *AzureAdApplication) GetSecretName() string {
	return i.Resource.Spec.SecretName
}

func (i *AzureAdApplication) GetOwnerReferences() []v1.OwnerReference {
	return i.Resource.GetOwnerReferences()
}

func (i *EntraID) IsRequestAuthEnabled() bool {
	return i != nil && i.RequestAuthentication != nil && i.RequestAuthentication.Enabled
}

func (i *EntraID) GetAuthSpec() *istiotypes.RequestAuthentication {
	if i != nil && i.RequestAuthentication != nil {
		return i.RequestAuthentication
	}
	return nil
}

func (i *EntraID) GetIdentityProvderName() IdentityProviderName {
	return EntraIDName
}

func (i *EntraID) GetProvidedSecretName() *string {
	if i != nil && i.RequestAuthentication != nil {
		return i.RequestAuthentication.SecretName
	}
	return nil
}

func (i *EntraID) GetPaths() []string {
	var paths []string
	if i.IsRequestAuthEnabled() {
		if i.RequestAuthentication.Paths != nil {
			paths = append(paths, *i.RequestAuthentication.Paths...)
		}
	}
	return paths
}

func (i *EntraID) GetIgnoredPaths() []string {
	var ignoredPaths []string
	if i.IsRequestAuthEnabled() {
		if i.RequestAuthentication.IgnorePaths != nil {
			ignoredPaths = append(ignoredPaths, *i.RequestAuthentication.IgnorePaths...)
		}
	}
	return ignoredPaths
}

func (i *EntraID) GetIssuerKey() string {
	return fmt.Sprintf("%s_OPENID_CONFIG_ISSUER", i.getSecretKeyPrefix())
}

func (i *EntraID) GetJwksKey() string {
	return fmt.Sprintf("%s_OPENID_CONFIG_JWKS_URI", i.getSecretKeyPrefix())
}

func (i *EntraID) GetClientIDKey() string {
	return fmt.Sprintf("%s_APP_CLIENT_ID", i.getSecretKeyPrefix())
}

func (i *EntraID) GetTokenLocation() string {
	if i != nil && i.RequestAuthentication != nil && i.RequestAuthentication.TokenLocation != nil {
		return *i.RequestAuthentication.TokenLocation
	}
	return "cookie"
}

func (i *EntraID) getSecretKeyPrefix() string {
	if i != nil && i.SecretKeyPrefix != nil {
		return *i.SecretKeyPrefix
	}
	return "AZURE"
}
