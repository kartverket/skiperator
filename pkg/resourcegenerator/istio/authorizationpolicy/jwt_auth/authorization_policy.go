package jwt_auth

import (
	"fmt"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/auth"
	"github.com/kartverket/skiperator/pkg/reconciliation"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/istio/authorizationpolicy"
	"github.com/kartverket/skiperator/pkg/util"
	securityv1api "istio.io/api/security/v1"
	typev1beta1 "istio.io/api/type/v1beta1"
	securityv1 "istio.io/client-go/pkg/apis/security/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func Generate(r reconciliation.Reconciliation) error {
	ctxLog := r.GetLogger()
	application, ok := r.GetSKIPObject().(*skiperatorv1alpha1.Application)
	if !ok {
		err := fmt.Errorf("failed to cast resource to application")
		ctxLog.Error(err, "Failed to generate JWT-auth AuthorizationPolicy")
		return err
	}
	ctxLog.Debug("Attempting to generate JWT-auth AuthorizationPolicy for application", "application", application.Name)

	if application.Spec.AuthorizationSettings != nil {
		// Do not create an AuthorizationPolicy if allowAll is set to true
		if application.Spec.AuthorizationSettings.AllowAll == true {
			return nil
		}
	}

	authConfigs := r.GetRequestAuthConfigs()
	if authConfigs == nil {
		ctxLog.Debug("No request auth configs provided for application. Skipping generating JWT-auth AuthorizationPolicy", "application", application.Name)
		return nil
	}

	if len(*authConfigs) > 0 {
		r.AddResource(
			getJwtValidationAuthPolicy(
				types.NamespacedName{
					Namespace: application.Namespace,
					Name:      application.Name + "-jwt-auth",
				},
				application.Name,
				*authConfigs,
			),
		)
	}
	ctxLog.Debug("Finished generating JWT-auth AuthorizationPolicy for application", "application", application.Name)
	return nil
}

func getJwtValidationAuthPolicy(namespacedName types.NamespacedName, applicationName string, authConfigs []auth.RequestAuthConfig) *securityv1.AuthorizationPolicy {
	var authPolicyRules []*securityv1api.Rule
	for _, authConfig := range authConfigs {
		var ruleTo securityv1api.Rule_To
		if len(authConfig.Paths) == 0 && len(authConfig.IgnorePaths) == 0 {
			ruleTo = securityv1api.Rule_To{
				Operation: &securityv1api.Operation{
					Paths: []string{"/*"},
				},
			}
		} else {
			ruleTo = securityv1api.Rule_To{
				Operation: &securityv1api.Operation{
					Paths:    authConfig.Paths,
					NotPaths: authConfig.IgnorePaths,
				},
			}
		}
		authPolicyRules = append(authPolicyRules, &securityv1api.Rule{
			To: []*securityv1api.Rule_To{
				&ruleTo,
			},
			When: []*securityv1api.Condition{
				{
					Key:    "request.auth.claims[iss]",
					Values: []string{authConfig.ProviderInfo.IssuerURI},
				},
				{
					Key:    "request.auth.claims[aud]",
					Values: []string{authConfig.ProviderInfo.ClientID},
				},
			},
			From: authorizationpolicy.GetGeneralFromRule(),
		})
	}
	return &securityv1.AuthorizationPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespacedName.Namespace,
			Name:      namespacedName.Name,
		},
		Spec: securityv1api.AuthorizationPolicy{
			Action: securityv1api.AuthorizationPolicy_ALLOW,
			Rules:  authPolicyRules,
			Selector: &typev1beta1.WorkloadSelector{
				MatchLabels: util.GetPodAppSelector(applicationName),
			},
		},
	}
}
