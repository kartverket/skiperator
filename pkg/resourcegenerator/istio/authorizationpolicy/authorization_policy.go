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
	defaultDenyAuthPolicy := getDefaultDenyPolicy(application, defaultDenyPaths)

	if application.Spec.AuthorizationSettings != nil {
		if application.Spec.AuthorizationSettings.AllowAll == true {
			return nil
		}
	}

	if application.Spec.AuthorizationSettings != nil {

		// As of now we only use one rule and one operation for all default denies. No need to loop over them all
		defaultDenyToOperation := defaultDenyAuthPolicy.Spec.Rules[0].To[0].Operation
		defaultDenyToOperation.NotPaths = nil

		if len(application.Spec.AuthorizationSettings.AllowList) > 0 {
			for _, endpoint := range application.Spec.AuthorizationSettings.AllowList {
				defaultDenyToOperation.NotPaths = append(defaultDenyToOperation.NotPaths, endpoint)
			}
		}
	}
	authConfig := r.GetAuthConfigs()
	if authConfig == nil {
		ctxLog.Debug("No auth config found. Skipping generating AuthorizationPolicy for JWT-validation for application", "application", application.Name)
	} else {
		ctxLog.Debug("Auth config found. Attempting to generate AuthorizationPolicy to validate JWT's for application", "application", application.Name)
		jwtValidationAuthPolicy, err := getJwtValidationAuthPolicy(application, *authConfig)
		if err != nil {
			ctxLog.Error(err, "Failed to generate AuthorizationPolicy to validate JWT's for application", "application", application.Name)
			return nil
		} else {
			ctxLog.Debug("Finished generating AuthorizationPolicy to validate JWT's for application", "application", application.Name)
			r.AddResource(jwtValidationAuthPolicy)
		}
	}

	ctxLog.Debug("Finished generating AuthorizationPolicy for application", "application", application.Name)
	r.AddResource(&defaultDenyAuthPolicy)

	return nil
}

func getJwtValidationAuthPolicy(application *skiperatorv1alpha1.Application, authConfigs []reconciliation.AuthConfig) (*securityv1.AuthorizationPolicy, error) {
	//TODO: Make common rule if ignorePaths are equal or if they are not present
	//TODO: Will AuthPolicy work if JWT is sent as BearerToken-Cookie
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

	return &securityv1.AuthorizationPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: application.Namespace,
			Name:      application.Name + "-require-jwt",
		},
		Spec: securityv1api.AuthorizationPolicy{
			Action: securityv1api.AuthorizationPolicy_ALLOW,
			Rules:  authPolicyRules,
		},
	}, nil
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

func getDefaultDenyPolicy(application *skiperatorv1alpha1.Application, denyPaths []string) securityv1.AuthorizationPolicy {
	return securityv1.AuthorizationPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: application.Namespace,
			Name:      application.Name + "-deny",
		},
		Spec: securityv1api.AuthorizationPolicy{
			Action: securityv1api.AuthorizationPolicy_DENY,
			Rules: []*securityv1api.Rule{
				{
					To: []*securityv1api.Rule_To{
						{
							Operation: &securityv1api.Operation{
								Paths: denyPaths,
							},
						},
					},
					From: getGeneralFromRule(),
				},
			},
			Selector: &typev1beta1.WorkloadSelector{
				MatchLabels: util.GetPodAppSelector(application.Name),
			},
		},
	}
}
