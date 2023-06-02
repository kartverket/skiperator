package applicationcontroller

import (
	"context"
	"github.com/kartverket/skiperator/pkg/util"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *SKIPJobReconciler) reconcileServiceAccount(ctx context.Context, skipJob *skiperatorv1alpha1.SKIPJob) (reconcile.Result, error) {

	serviceAccount := corev1.ServiceAccount{ObjectMeta: metav1.ObjectMeta{Namespace: skipJob.Namespace, Name: skipJob.Name}}

	_, err := ctrlutil.CreateOrPatch(ctx, r.GetClient(), &serviceAccount, func() error {
		// Set application as owner of the sidecar
		err := ctrlutil.SetControllerReference(skipJob, &serviceAccount, r.GetScheme())
		if err != nil {
			return err
		}

		util.SetCommonAnnotations(&serviceAccount)

		return nil
	})

	return reconcile.Result{}, err
}
