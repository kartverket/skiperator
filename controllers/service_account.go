package controllers

import (
	"context"

	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *ApplicationReconciler) reconcileServiceAccount(ctx context.Context, application *skiperatorv1alpha1.Application) (reconcile.Result, error) {
	controllerName := "ServiceAccount"
	r.ManageControllerStatus(ctx, application, controllerName, skiperatorv1alpha1.PROGRESSING)

	serviceAccount := corev1.ServiceAccount{ObjectMeta: metav1.ObjectMeta{Namespace: application.Namespace, Name: application.Name}}
	_, err := ctrlutil.CreateOrPatch(ctx, r.GetClient(), &serviceAccount, func() error {
		// Set application as owner of the sidecar
		err := ctrlutil.SetControllerReference(application, &serviceAccount, r.GetScheme())
		if err != nil {
			r.ManageControllerStatusError(ctx, application, controllerName, err)
			return err
		}

		return nil
	})

	r.ManageControllerOutcome(ctx, application, controllerName, skiperatorv1alpha1.SYNCED, err)

	return reconcile.Result{}, err
}
