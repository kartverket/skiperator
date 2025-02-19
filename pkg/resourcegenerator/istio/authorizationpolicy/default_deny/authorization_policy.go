package default_deny

import (
	"fmt"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/reconciliation"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/istio/authorizationpolicy"
	securityv1api "istio.io/api/security/v1"
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
	ctxLog.Debug("Attempting to generate default deny AuthorizationPolicy for application", "application", application.Name)

	if application.Spec.AuthorizationSettings != nil {
		// Do not create an AuthorizationPolicy if allowAll is set to true
		if application.Spec.AuthorizationSettings.AllowAll == true {
			return nil
		}
	}

	r.AddResource(
		authorizationpolicy.GetAuthPolicy(
			types.NamespacedName{
				Name:      application.Name + "-default-deny",
				Namespace: application.Namespace,
			},
			application.Name,
			securityv1api.AuthorizationPolicy_DENY,
			[]string{authorizationpolicy.DefaultDenyPath},
			[]string{},
		),
	)

	ctxLog.Debug("Finished generating default AuthorizationPolicy for application", "application", application.Name)
	return nil
}
