package default_deny

import (
	"fmt"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/reconciliation"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/istio/authorizationpolicy"
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
	application, ok := r.GetSKIPObject().(*skiperatorv1alpha1.Application)
	if !ok {
		err := fmt.Errorf("failed to cast resource to application")
		ctxLog.Error(err, "Failed to generate default AuthorizationPolicy")
		return err
	}
	ctxLog.Debug("Attempting to generate default AuthorizationPolicy for application", "application", application.Name)

	if application.Spec.AuthorizationSettings != nil {
		// Do not create an AuthorizationPolicy if allowAll is set to true
		if application.Spec.AuthorizationSettings.AllowAll == true {
			return nil
		}
	}

	var allowedPaths []string
	if application.Spec.AuthorizationSettings != nil {
		allowedPaths = append(allowedPaths, application.Spec.AuthorizationSettings.AllowList...)
	}

	// Generate an AuthorizationPolicy that allows requests to the list of paths in allowPaths
	if len(allowedPaths) > 0 {
		r.AddResource(
			getAuthPolicy(
				types.NamespacedName{
					Name:      application.Name + "-allow-paths",
					Namespace: application.Namespace,
				},
				application.Name,
				securityv1api.AuthorizationPolicy_ALLOW,
				allowedPaths,
				[]string{},
			),
		)
	} else {
		r.AddResource(
			getAuthPolicy(
				types.NamespacedName{
					Name:      application.Name + "-default-deny",
					Namespace: application.Namespace,
				},
				application.Name,
				securityv1api.AuthorizationPolicy_DENY,
				[]string{authorizationpolicy.DefaultDenyPath},
				allowedPaths,
			),
		)
	}
	ctxLog.Debug("Finished generating default AuthorizationPolicy for application", "application", application.Name)
	return nil
}

func getAuthPolicy(namespacedName types.NamespacedName, applicationName string, action v1beta1.AuthorizationPolicy_Action, paths []string, notPaths []string) *securityv1.AuthorizationPolicy {
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
					From: authorizationpolicy.GetGeneralFromRule(),
				},
			},
			Selector: &typev1beta1.WorkloadSelector{
				MatchLabels: util.GetPodAppSelector(applicationName),
			},
		},
	}
}
