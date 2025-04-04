package default_deny

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
	"strings"
)

func Generate(r reconciliation.Reconciliation) error {
	ctxLog := r.GetLogger()
	application, ok := r.GetSKIPObject().(*skiperatorv1alpha1.Application)
	if !ok {
		err := fmt.Errorf("failed to cast resource to application")
		ctxLog.Error(err, "Failed to generate default AuthorizationPolicy")
		return err
	}
	ctxLog.Debug("Attempting to generate default deny AuthorizationPolicy for application", "application", application.Name)

	if application.Spec.AuthorizationSettings != nil {
		// Do not create an AuthorizationPolicy if allowAll is set to true
		if application.Spec.AuthorizationSettings.AllowAll == true {
			return nil
		}
	}

	defaultDenyPath := authorizationpolicy.DefaultDenyPath
	if application.Spec.IsRequestAuthEnabled() && r.GetAuthConfigs() == nil {
		defaultDenyPath = "*"
		ctxLog.Debug("No auth config provided. Defaults to deny-all AuthorizationPolicy for application", "application", application.Name)
	}

	var notPaths []string
	trimmedPrefix := strings.TrimSuffix(authorizationpolicy.DefaultDenyPath, "*")
	if application.Spec.AuthorizationSettings != nil {
		for _, path := range application.Spec.AuthorizationSettings.AllowList {
			if strings.HasPrefix(path, trimmedPrefix) {
				notPaths = append(notPaths, path)
			}
		}
	}
	for _, path := range r.GetAuthConfigs().GetAllPaths() {
		if strings.HasPrefix(path, trimmedPrefix) {
			notPaths = append(notPaths, path)
		}
	}

	r.AddResource(
		&securityv1.AuthorizationPolicy{
			ObjectMeta: metav1.ObjectMeta{
				Name:      application.Name + "-default-deny",
				Namespace: application.Namespace,
			},
			Spec: securityv1api.AuthorizationPolicy{
				Action: securityv1api.AuthorizationPolicy_DENY,
				Rules: []*securityv1api.Rule{
					{
						To: []*securityv1api.Rule_To{
							{
								Operation: &securityv1api.Operation{
									Paths:    []string{defaultDenyPath},
									NotPaths: notPaths,
								},
							},
						},
						From: authorizationpolicy.GetGeneralFromRule(),
					},
				},
				Selector: &typev1beta1.WorkloadSelector{
					MatchLabels: util.GetPodAppSelector(application.Name),
				},
			},
		},
	)

	ctxLog.Debug("Finished generating default AuthorizationPolicy for application", "application", application.Name)
	return nil
}
