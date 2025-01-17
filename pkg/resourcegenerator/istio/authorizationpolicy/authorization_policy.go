package authorizationpolicy

import (
	"fmt"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/reconciliation"
	"github.com/kartverket/skiperator/pkg/util"
	securityv1api "istio.io/api/security/v1"
	typev1beta1 "istio.io/api/type/v1beta1"
	securityv1 "istio.io/client-go/pkg/apis/security/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	defaultDenyPaths := []string{
		"/actuator*",
	}
	authPolicy := getAuthorizationPolicy(application, defaultDenyPaths, r.GetAuthConfigs())

	ctxLog.Debug("Finished generating AuthorizationPolicy for application", "application", application.Name)
	if authPolicy != nil {
		r.AddResource(authPolicy)
	}

	return nil
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

func getAuthorizationPolicy(application *skiperatorv1alpha1.Application, denyPaths []string, authConfigs *[]reconciliation.AuthConfig) *securityv1.AuthorizationPolicy {
	authPolicyRules := []*securityv1api.Rule{
		{
			To:   []*securityv1api.Rule_To{},
			From: getGeneralFromRule(),
		},
	}

	if application.Spec.AuthorizationSettings != nil {
		if application.Spec.AuthorizationSettings.AllowAll == true {
			return nil
		}
		// As of now we only use one rule and one operation for all default denies. No need to loop over them all
		if len(application.Spec.AuthorizationSettings.AllowList) > 0 {
			operation := authPolicyRules[0].To[0].Operation
			for _, endpoint := range application.Spec.AuthorizationSettings.AllowList {
				operation.Paths = append(operation.Paths, endpoint)
			}
			authPolicyRules[0].To[0].Operation = operation
		} else {
			authPolicyRules[0].To[0].Operation.NotPaths = denyPaths
		}
	} else {
		authPolicyRules[0].To[0].Operation.NotPaths = denyPaths
	}

	if authConfigs != nil {
		authPolicyRules = append(authPolicyRules, getJwtValidationRule(*authConfigs)...)
	}

	return &securityv1.AuthorizationPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: application.Namespace,
			Name:      application.Name + "-auth-policy",
		},
		Spec: securityv1api.AuthorizationPolicy{
			Action: securityv1api.AuthorizationPolicy_ALLOW,
			Rules:  authPolicyRules,
			Selector: &typev1beta1.WorkloadSelector{
				MatchLabels: util.GetPodAppSelector(application.Name),
			},
		},
	}
}

func getJwtValidationRule(authConfigs []reconciliation.AuthConfig) []*securityv1api.Rule {
	authPolicyRules := make([]*securityv1api.Rule, 0)
	for _, authConfig := range authConfigs {
		ruleTo := &securityv1api.Rule_To{}
		if authConfig.NotPaths != nil {
			ruleTo.Operation = &securityv1api.Operation{
				NotPaths: *authConfig.NotPaths,
			}
		} else {
			ruleTo.Operation = &securityv1api.Operation{
				Paths: []string{"*"},
			}
		}
		authPolicyRules = append(authPolicyRules, &securityv1api.Rule{
			To: []*securityv1api.Rule_To{ruleTo},
			When: []*securityv1api.Condition{
				{
					Key:    "request.auth.claims[iss]",
					Values: []string{authConfig.ProviderURIs.IssuerURI},
				},
			},
		})
	}
	return authPolicyRules
}
