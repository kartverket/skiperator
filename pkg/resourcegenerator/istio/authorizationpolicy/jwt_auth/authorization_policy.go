package jwt_auth

import (
	"fmt"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
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

	authConfigs := r.GetAuthConfigs()
	allowedPaths := authConfigs.GetAllowedPaths(application.Spec.AuthorizationSettings)
	if authConfigs != nil {
		if len(*authConfigs) > 0 {
			r.AddResource(
				getJwtValidationAuthPolicy(
					types.NamespacedName{
						Namespace: application.Namespace,
						Name:      application.Name + "-jwt-auth",
					},
					application.Name,
					*authConfigs,
					allowedPaths,
					[]string{authorizationpolicy.DefaultDenyPath},
				),
			)
		}
	}
	ctxLog.Debug("Finished generating JWT-auth AuthorizationPolicy for application", "application", application.Name)
	return nil
}

func getJwtValidationAuthPolicy(namespacedName types.NamespacedName, applicationName string, authConfigs []reconciliation.AuthConfig, allowPaths []string, denyPaths []string) *securityv1.AuthorizationPolicy {
	var authPolicyRules []*securityv1api.Rule

	notPaths := allowPaths
	notPaths = append(allowPaths, denyPaths...)
	for _, authConfig := range authConfigs {
		authPolicyRules = append(authPolicyRules, &securityv1api.Rule{
			To: []*securityv1api.Rule_To{
				{
					Operation: &securityv1api.Operation{
						NotPaths: notPaths,
					},
				},
			},
			When: []*securityv1api.Condition{
				{
					Key:    "request.auth.claims[iss]",
					Values: []string{authConfig.ProviderURIs.IssuerURI},
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
