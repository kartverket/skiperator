package authorizationpolicy

import (
	"fmt"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/reconciliation"
	"github.com/kartverket/skiperator/pkg/util"
	securityv1api "istio.io/api/security/v1"
	"istio.io/api/security/v1beta1"
	typev1beta1 "istio.io/api/type/v1beta1"
	securityv1 "istio.io/client-go/pkg/apis/security/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func Generate(r reconciliation.Reconciliation) error {
	ctxLog := r.GetLogger()
	if r.GetType() != reconciliation.ApplicationType {
		return fmt.Errorf("unsupported type %s in AuthorizationPolicy", r.GetType())
	}
	application, ok := r.GetSKIPObject().(*skiperatorv1alpha1.Application)
	if !ok {
		err := fmt.Errorf("failed to cast resource to application")
		ctxLog.Error(err, "Failed to generate AuthorizationPolicy")
		return err
	}
	ctxLog.Debug("Attempting to generate AuthorizationPolicy for application", "application", application.Name)

	if application.Spec.AuthorizationSettings != nil {
		// Do not create an AuthorizationPolicy if allowAll is set to true
		if application.Spec.AuthorizationSettings.AllowAll == true {
			return nil
		}
	}
	defaultDenyPath := []string{"/actuator*"}

	allowPaths := []string{}
	if application.Spec.AuthorizationSettings != nil {
		if application.Spec.AuthorizationSettings.AllowList != nil {
			if len(application.Spec.AuthorizationSettings.AllowList) > 0 {
				allowPaths = application.Spec.AuthorizationSettings.AllowList
			}
		}
	}
	authConfigs := r.GetAuthConfigs()
	if authConfigs != nil {
		if len(*authConfigs) > 0 {
			for _, authConfig := range *authConfigs {
				if authConfig.NotPaths != nil {
					allowPaths = append(allowPaths, *authConfig.NotPaths...)
				}
			}
			r.AddResource(
				getJwtValidationAuthPolicy(
					types.NamespacedName{
						Namespace: application.Namespace,
						Name:      application.Name + "-jwt-auth",
					},
					application.Name,
					*authConfigs,
					allowPaths,
					defaultDenyPath,
				),
			)
		}
	}

	// Generate an AuthorizationPolicy that allows requests to the list of paths in allowPaths
	if len(allowPaths) > 0 {
		r.AddResource(
			getGeneralAuthPolicy(
				types.NamespacedName{
					Name:      application.Name + "-allow-paths",
					Namespace: application.Namespace,
				},
				application.Name,
				securityv1api.AuthorizationPolicy_ALLOW,
				allowPaths,
				[]string{},
			),
		)
	} else {
		r.AddResource(
			getGeneralAuthPolicy(
				types.NamespacedName{
					Name:      application.Name + "-default-deny",
					Namespace: application.Namespace,
				},
				application.Name,
				securityv1api.AuthorizationPolicy_DENY,
				defaultDenyPath,
				allowPaths,
			),
		)
	}
	ctxLog.Debug("Finished generating AuthorizationPolicy for application", "application", application.Name)
	return nil
}

func getGeneralAuthPolicy(namespacedName types.NamespacedName, applicationName string, action v1beta1.AuthorizationPolicy_Action, paths []string, notPaths []string) *securityv1.AuthorizationPolicy {
	return &securityv1.AuthorizationPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespacedName.Namespace,
			Name:      namespacedName.Name,
		},
		Spec: securityv1api.AuthorizationPolicy{
			Action: action,
			Rules: []*securityv1api.Rule{
				{
					To: []*securityv1api.Rule_To{
						{
							Operation: &securityv1api.Operation{
								Paths:    paths,
								NotPaths: notPaths,
							},
						},
					},
					From: getGeneralFromRule(),
				},
			},
			Selector: &typev1beta1.WorkloadSelector{
				MatchLabels: util.GetPodAppSelector(applicationName),
			},
		},
	}
}

func getGeneralFromRule() []*securityv1api.Rule_From {
	return []*securityv1api.Rule_From{
		{
			Source: &securityv1api.Source{
				Namespaces: []string{"istio-gateways"},
			},
		},
	}
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
			From: getGeneralFromRule(),
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
