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
	application.FillDefaults()
	controllerName := "serviceaccount"
	controllerMessageName := "ServiceAccount"
	r.ManageControllerStatus(ctx, application, controllerName, skiperatorv1alpha1.Status{Message: controllerMessageName + " starting reconciliation", Status: skiperatorv1alpha1.PROGRESSING})

	serviceAccount := corev1.ServiceAccount{ObjectMeta: metav1.ObjectMeta{Namespace: application.Namespace, Name: application.Name}}
	_, err := ctrlutil.CreateOrPatch(ctx, r.GetClient(), &serviceAccount, func() error {
		// Set application as owner of the sidecar
		err := ctrlutil.SetControllerReference(application, &serviceAccount, r.GetScheme())
		if err != nil {
			r.ManageControllerStatus(ctx, application, controllerName, skiperatorv1alpha1.Status{Message: controllerMessageName + " encountered error: " + err.Error(), Status: skiperatorv1alpha1.ERROR})
			return err
		}

		return nil
	})

	if err != nil {
		r.ManageControllerStatus(ctx, application, controllerName, skiperatorv1alpha1.Status{Message: controllerMessageName + " encountered error: " + err.Error(), Status: skiperatorv1alpha1.ERROR})
	} else {
		r.ManageControllerStatus(ctx, application, controllerName, skiperatorv1alpha1.Status{Message: controllerMessageName + " synced", Status: skiperatorv1alpha1.SYNCED})
	}

	return reconcile.Result{}, err
}
