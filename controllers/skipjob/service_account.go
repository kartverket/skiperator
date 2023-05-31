package applicationcontroller

import (
	"context"
	applicationcontroller "github.com/kartverket/skiperator/controllers/application"

	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *SKIPJobReconciler) reconcileServiceAccount(ctx context.Context, skipJob *skiperatorv1alpha1.SKIPJob) (reconcile.Result, error) {

	serviceAccount := corev1.ServiceAccount{ObjectMeta: metav1.ObjectMeta{Namespace: skipJob.Namespace, Name: skipJob.Name}}
	err := applicationcontroller.CreateOrPatchServiceAccount(ctx, serviceAccount, skipJob, r.GetClient(), r.GetScheme())

	return reconcile.Result{}, err
}
