package requestauthentication

import (
	"fmt"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/api/v1alpha1/istiotypes"
	"github.com/kartverket/skiperator/pkg/jwtAuth"
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

	authConfig := r.GetAuthConfigs()

	if authConfig == nil {
		ctxLog.Debug("No RequestAuthentication to generate. No jwtAuth config provided for", "application", application.Name)
	} else {
		requestAuthentication := getRequestAuthentication(application, *authConfig)
		r.AddResource(&requestAuthentication)
		ctxLog.Debug("Finished generating RequestAuthentication for application", "application", application.Name)
	}
	return nil
}

func getRequestAuthentication(application *skiperatorv1alpha1.Application, authConfigs []jwtAuth.AuthConfig) securityv1.RequestAuthentication {
	jwtRules := make([]*v1beta1.JWTRule, len(authConfigs))
	for i, config := range authConfigs {
		switch config.ProviderURIs.Provider {
		case jwtAuth.ID_PORTEN:
			jwtRules[i] = getJWTRule(application.Spec.IDPorten.Authentication, config.ProviderURIs)
		case jwtAuth.MASKINPORTEN:
			jwtRules[i] = getJWTRule(application.Spec.Maskinporten.Authentication, config.ProviderURIs)
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

func getJWTRule(authentication *istiotypes.Authentication, providerURIs jwtAuth.ProviderURIs) *v1beta1.JWTRule {
	var jwtRule = v1beta1.JWTRule{
		ForwardOriginalToken: authentication.ForwardOriginalToken,
	}
	if authentication.TokenLocation == "cookie" {
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

	jwtRule.Issuer = providerURIs.IssuerURI
	jwtRule.JwksUri = providerURIs.JwksURI
	jwtRule.Audiences = []string{providerURIs.ClientID}

	return &jwtRule
}
