package webhook

import (
	ctrl "sigs.k8s.io/controller-runtime"

	v1beta1 "github.com/kartverket/skiperator/api/v1beta1"
)

// nolint:unused
// log is for logging in this package.
//var SKIPJobLog = logf.Log.WithName("svartskjaif-resource")

// SetupSvartSkjaifWebhookWithManager registers the webhook for SvartSkjaif in the manager.
func SetupSkipJobWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&v1beta1.SKIPJob{}).
		Complete()
}
