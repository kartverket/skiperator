package jwt_auth

import (
	"fmt"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/api/v1alpha1/istiotypes"
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

	authConfigs := r.GetAuthConfigs()
	if authConfigs == nil {
		ctxLog.Debug("No auth configs provided for application. Skipping generating JWT-auth AuthorizationPolicy", "application", application.Name)
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

func getJwtValidationAuthPolicy(namespacedName types.NamespacedName, applicationName string, authConfigs []auth.AuthConfig) *securityv1.AuthorizationPolicy {
	var authPolicyRules []*securityv1api.Rule
	for _, authConfig := range authConfigs {
		baseConditions := getBaseConditions(authConfig)
		if len(authConfig.AuthRules)+len(authConfig.IgnoreAuthRules) == 0 {
			authPolicyRules = append(authPolicyRules, &securityv1api.Rule{
				To: []*securityv1api.Rule_To{
					{
						Operation: &securityv1api.Operation{
							Paths: []string{"*"},
						},
					},
				},
				When: baseConditions,
			})
		} else {
			// The first rule ensures requestMatchers not covered in auth configs require valid JWT.
			// This is the same as the diff between all possible requests and the list of all requestMatchers in authRules + ignoreAuth.
			authPolicyRules = append(authPolicyRules, &securityv1api.Rule{
				To: authorizationpolicy.GetApiSurfaceDiffAsRuleToList(
					istiotypes.RequestMatchers{
						{
							Paths: []string{"*"},
						},
					},
					append(authConfig.AuthRules.GetRequestMatchers(), authConfig.IgnoreAuthRules...),
				),
				From: authorizationpolicy.GetGeneralFromRule(),
				When: baseConditions,
			})

			for _, rule := range authConfig.AuthRules {
				authPolicyRules = append(
					authPolicyRules,
					// Apply rule to enforce rules specified by user
					&securityv1api.Rule{
						To: []*securityv1api.Rule_To{
							{
								Operation: &securityv1api.Operation{
									Paths:   rule.Paths,
									Methods: rule.Methods,
								},
							},
						},
						From: authorizationpolicy.GetGeneralFromRule(),
						When: append(baseConditions, getAuthPolicyRuleConditions(rule.When)...),
					},
				)
			}
		}
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

func getAuthPolicyRuleConditions(conditions []istiotypes.Condition) []*securityv1api.Condition {
	var istioConditions []*securityv1api.Condition
	for _, whenCondition := range conditions {
		istioConditions = append(istioConditions, &securityv1api.Condition{
			Key:    fmt.Sprintf("request.auth.claims[%s]", whenCondition.Claim),
			Values: whenCondition.Values,
		})
	}
	return istioConditions
}

func getBaseConditions(authConfig auth.AuthConfig) []*securityv1api.Condition {
	conditions := []*securityv1api.Condition{
		{
			Key:    "request.auth.claims[iss]",
			Values: []string{authConfig.ProviderInfo.IssuerURI},
		},
		{
			Key:    "request.auth.claims[aud]",
			Values: []string{authConfig.ProviderInfo.ClientID},
		},
	}
	if len(authConfig.AcceptedResources) == 0 {
		return conditions
	}
	return append(conditions, &securityv1api.Condition{
		Key:    "request.auth.claims[aud]",
		Values: authConfig.AcceptedResources,
	})
}
