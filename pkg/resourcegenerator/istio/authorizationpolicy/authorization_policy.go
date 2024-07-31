package authorizationpolicy

import (
	"fmt"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/reconciliation"
	"github.com/kartverket/skiperator/pkg/util"
	securityv1beta1api "istio.io/api/security/v1beta1"
	typev1beta1 "istio.io/api/type/v1beta1"
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func Generate(r reconciliation.Reconciliation) error {
	ctxLog := r.GetLogger()
	if r.GetType() != reconciliation.ApplicationType {
		return fmt.Errorf("unsupported type %s in peer authentication", r.GetType())
	}
	application, ok := r.GetSKIPObject().(*skiperatorv1alpha1.Application)
	if !ok {
		err := fmt.Errorf("failed to cast resource to application")
		ctxLog.Error(err, "Failed to generate peer authentication")
		return err
	}
	ctxLog.Debug("Attempting to generate network policy for application", "application", application.Name)

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

	var obj client.Object = &defaultDenyAuthPolicy
	r.AddResource(obj)

	return nil
}

func getGeneralFromRule() []*securityv1beta1api.Rule_From {
	return []*securityv1beta1api.Rule_From{
		{
			Source: &securityv1beta1api.Source{
				Namespaces: []string{"istio-gateways"},
			},
		},
	}
}

func getDefaultDenyPolicy(application *skiperatorv1alpha1.Application, denyPaths []string) securityv1beta1.AuthorizationPolicy {
	return securityv1beta1.AuthorizationPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: application.Namespace,
			Name:      application.Name + "-deny",
		},
		Spec: securityv1beta1api.AuthorizationPolicy{
			Action: securityv1beta1api.AuthorizationPolicy_DENY,
			Rules: []*securityv1beta1api.Rule{
				{
					To: []*securityv1beta1api.Rule_To{
						{
							Operation: &securityv1beta1api.Operation{
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
