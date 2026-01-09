package requestauthentication

import (
	"fmt"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/v2/api/v1alpha1"
	"github.com/kartverket/skiperator/v2/pkg/auth"
	"github.com/kartverket/skiperator/v2/pkg/reconciliation"
	"github.com/kartverket/skiperator/v2/pkg/util"
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

	authConfigs := r.GetAuthConfigs()

	if authConfigs == nil {
		ctxLog.Debug("No auth configs provided for application. Skipping generating RequestAuthentication", "application", application.Name)
		return nil
	}
	requestAuthentication := getRequestAuthentication(application, *authConfigs)
	r.AddResource(&requestAuthentication)
	ctxLog.Debug("Finished generating RequestAuthentication for application", "application", application.Name)
	return nil
}

func getRequestAuthentication(application *skiperatorv1alpha1.Application, authConfigs []auth.AuthConfig) securityv1.RequestAuthentication {
	jwtRules := make([]*v1beta1.JWTRule, len(authConfigs))
	for i, config := range authConfigs {
		jwtRules[i] = getJWTRule(config)
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

func getJWTRule(authConfig auth.AuthConfig) *v1beta1.JWTRule {
	var jwtRule = v1beta1.JWTRule{
		ForwardOriginalToken: authConfig.Spec.ForwardJwt,
	}
	if authConfig.TokenLocation == "cookie" {
		jwtRule.FromCookies = []string{"BearerToken"}
	}
	if authConfig.Spec.OutputClaimToHeaders != nil {
		claimsToHeaders := make([]*v1beta1.ClaimToHeader, len(*authConfig.Spec.OutputClaimToHeaders))
		for i, claimToHeader := range *authConfig.Spec.OutputClaimToHeaders {
			claimsToHeaders[i] = &v1beta1.ClaimToHeader{
				Header: claimToHeader.Header,
				Claim:  claimToHeader.Claim,
			}
		}
		jwtRule.OutputClaimToHeaders = claimsToHeaders
	}

	jwtRule.Issuer = authConfig.ProviderInfo.IssuerURI
	jwtRule.JwksUri = authConfig.ProviderInfo.JwksURI
	jwtRule.Audiences = []string{authConfig.ProviderInfo.ClientID}

	return &jwtRule
}
