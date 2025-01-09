package requestauthentication

import (
	"fmt"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/api/v1alpha1/istiotypes"
	"github.com/kartverket/skiperator/pkg/reconciliation"
	"github.com/kartverket/skiperator/pkg/util"
	securityv1api "istio.io/api/security/v1"
	"istio.io/api/security/v1beta1"
	typev1beta1 "istio.io/api/type/v1beta1"
	securityv1 "istio.io/client-go/pkg/apis/security/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Generate(r reconciliation.Reconciliation) error {
	ctxLog := r.GetLogger()
	if r.GetType() != reconciliation.ApplicationType {
		return fmt.Errorf("unsupported type %s in RequestAuthentication", r.GetType())
	}
	application, ok := r.GetSKIPObject().(*skiperatorv1alpha1.Application)
	if !ok {
		err := fmt.Errorf("failed to cast resource to application")
		ctxLog.Error(err, "Failed to generate RequestAuthentication")
		return err
	}
	ctxLog.Debug("Attempting to generate RequestAuthentication for application", "application", application.Name)
	providers := getIdentityProvidersWithAuthenticationEnabled(application)
	if providers != nil {
		requestAuthentication := getRequestAuthentication(application, *providers)
		ctxLog.Debug("Finished generating RequestAuthentication for application", "application", application.Name)
		r.AddResource(&requestAuthentication)
	}
	return nil
}

func getIdentityProvidersWithAuthenticationEnabled(application *skiperatorv1alpha1.Application) *[]IdentityProvider {
	var providers []IdentityProvider
	if application.Spec.IDPorten != nil && application.Spec.IDPorten.Enabled && application.Spec.IDPorten.Authentication != nil && application.Spec.IDPorten.Authentication.Enabled {
		providers = append(providers, ID_PORTEN)
	}
	if application.Spec.Maskinporten != nil && application.Spec.Maskinporten.Enabled && application.Spec.Maskinporten.Authentication != nil && application.Spec.Maskinporten.Authentication.Enabled == true {
		providers = append(providers, MASKINPORTEN)
	}
	if len(providers) > 0 {
		return &providers
	} else {
		return nil
	}
}

func getRequestAuthentication(application *skiperatorv1alpha1.Application, providers []IdentityProvider) securityv1.RequestAuthentication {
	jwtRules := make([]*v1beta1.JWTRule, len(providers))
	for i, provider := range providers {
		if provider == ID_PORTEN {
			jwtRules[i] = getJWTRule(application.Spec.IDPorten.Authentication, provider)
		} else if provider == MASKINPORTEN {
			jwtRules[i] = getJWTRule(application.Spec.Maskinporten.Authentication, provider)
		}
	}
	return securityv1.RequestAuthentication{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: application.Namespace,
			Name:      application.Name + "-jwt-authn",
		},
		Spec: securityv1api.RequestAuthentication{
			Selector: &typev1beta1.WorkloadSelector{
				MatchLabels: util.GetPodAppSelector(application.Name),
			},
			JwtRules: jwtRules,
		},
	}
}

func getJWTRule(authentication *istiotypes.Authentication, provider IdentityProvider) *v1beta1.JWTRule {
	var forwardOriginalToken = true
	if authentication.ForwardOriginalToken != nil {
		forwardOriginalToken = *authentication.ForwardOriginalToken
	}
	var jwtRule = v1beta1.JWTRule{
		Audiences:            []string{}, //TODO: Retrieve audience
		ForwardOriginalToken: forwardOriginalToken,
	}
	if authentication.TokenLocation != nil && *authentication.TokenLocation == cookie {
		jwtRule.FromCookies = []string{"BearerToken"}
	}

	if authentication.OutputClaimToHeaders != nil {
		claimsToHeaders := make([]*v1beta1.ClaimToHeader, len(*authentication.OutputClaimToHeaders))
		for i, claimToHeader := range *authentication.OutputClaimToHeaders {
			claimsToHeaders[i] = &v1beta1.ClaimToHeader{
				Header: claimToHeader.Header,
				Claim:  claimToHeader.Claim,
			}
		}
		jwtRule.OutputClaimToHeaders = claimsToHeaders
	}

	switch provider {
	case ID_PORTEN:
		jwtRule.Issuer = "https://idporten.no"
		jwtRule.JwksUri = "https://idporten.no/jwks.json"
		if authentication.TokenLocation == nil {
			jwtRule.FromCookies = []string{"BearerToken"}
		}

	case MASKINPORTEN:
		jwtRule.Issuer = "https://maskinporten.no"
		jwtRule.JwksUri = "https://maskinporten.no/jwk"
	}
	return &jwtRule
}

type IdentityProvider string

var (
	MASKINPORTEN IdentityProvider = "MASKINPORTEN"
	ID_PORTEN    IdentityProvider = "ID_PORTEN"
)

var (
	header = "header"
	cookie = "cookie"
)
