package allow

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
)

func Generate(r reconciliation.Reconciliation) error {
	ctxLog := r.GetLogger()
	application, ok := r.GetSKIPObject().(*skiperatorv1alpha1.Application)
	if !ok {
		err := fmt.Errorf("failed to cast resource to application")
		ctxLog.Error(err, "Failed to generate default AuthorizationPolicy")
		return err
	}
	ctxLog.Debug("Attempting to generate allow AuthorizationPolicy for application", "application", application.Name)

	if application.Spec.AuthorizationSettings != nil {
		// Do not create an AuthorizationPolicy if allowAll is set to true
		if application.Spec.AuthorizationSettings.AllowAll == true {
			return nil
		}
	}

	authConfigs := r.GetAuthConfigs()
	if authConfigs == nil {
		ctxLog.Debug("No auth configs provided for application. Skipping generating allow-paths AuthorizationPolicy", "application", application.Name)
		return nil
	}

	var authPolicyRules []*securityv1api.Rule

	// Include ignored paths from auth config as they should be accessible without authentication
	allowedPaths := r.GetAuthConfigs().GetIgnoredPaths()
	if application.Spec.AuthorizationSettings != nil {
		allowedPaths = append(allowedPaths, application.Spec.AuthorizationSettings.AllowList...)
	}

	if len(allowedPaths) > 0 {
		authPolicyRules = append(authPolicyRules, &securityv1api.Rule{
			To: []*securityv1api.Rule_To{
				{
					Operation: &securityv1api.Operation{
						Paths: allowedPaths,
					},
				},
			},
			From: authorizationpolicy.GetGeneralFromRule(),
		})
	}

	var allowedPrincipals []string
	if application.Spec.AccessPolicy != nil && application.Spec.AccessPolicy.Inbound != nil {
		for _, rule := range application.Spec.AccessPolicy.Inbound.Rules {
			allowedPrincipals = append(allowedPrincipals, rule.ToPrincipal(application.Namespace))
		}
	}

	if len(allowedPrincipals) > 0 {
		authPolicyRules = append(authPolicyRules, &securityv1api.Rule{
			From: []*securityv1api.Rule_From{
				{
					Source: &securityv1api.Source{
						Principals: allowedPrincipals,
					},
				},
			},
		})
	}

	// Generate an AuthorizationPolicy that allows requests to the list of paths in allowPaths
	if len(authPolicyRules) > 0 && len(*authConfigs) > 0 {
		r.AddResource(
			&securityv1.AuthorizationPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      application.Name + "-allow-paths",
					Namespace: application.Namespace,
				},
				Spec: securityv1api.AuthorizationPolicy{
					Action: securityv1api.AuthorizationPolicy_ALLOW,
					Rules:  authPolicyRules,
					Selector: &typev1beta1.WorkloadSelector{
						MatchLabels: util.GetPodAppSelector(application.Name),
					},
				},
			},
		)
	}
	ctxLog.Debug("Finished generating allow AuthorizationPolicy for application", "application", application.Name)
	return nil
}
