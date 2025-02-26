package allow

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
	ctxLog.Debug("Attempting to generate allow AuthorizationPolicy for application", "application", application.Name)

	if application.Spec.AuthorizationSettings != nil {
		// Do not create an AuthorizationPolicy if allowAll is set to true
		if application.Spec.AuthorizationSettings.AllowAll == true {
			return nil
		}
	}

	var allowedPaths []string
	if application.Spec.AuthorizationSettings != nil {
		allowedPaths = append(allowedPaths, application.Spec.AuthorizationSettings.AllowList...)
		// Include ignored paths from auth config as they should be accessible without authentication
		allowedPaths = append(allowedPaths, r.GetAuthConfigs().GetIgnoredPaths()...)
	}

	// Generate an AuthorizationPolicy that allows requests to the list of paths in allowPaths
	if len(allowedPaths) > 0 {
		r.AddResource(
			authorizationpolicy.GetAuthPolicy(
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
	}
	ctxLog.Debug("Finished generating allow AuthorizationPolicy for application", "application", application.Name)
	return nil
}
