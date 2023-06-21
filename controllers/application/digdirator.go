package applicationcontroller

import (
	"context"

	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/util"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *ApplicationReconciler) reconcileDigdirator(ctx context.Context, application *skiperatorv1alpha1.Application) (reconcile.Result, error) {
	controllerName := "Digdirator"
	r.SetControllerProgressing(ctx, application, controllerName)

	// if !application.Spec.Digdirator {
	// 	return reconcile.Result{}, nil
	// }

	digdirator := corev1.ServiceAccount{ObjectMeta: metav1.ObjectMeta{Namespace: application.Namespace, Name: application.Name}}

	_, err := ctrlutil.CreateOrPatch(ctx, r.GetClient(), &digdirator, func() error {
		// Set application as owner of the sidecar
		err := ctrlutil.SetControllerReference(application, &digdirator, r.GetScheme())
		if err != nil {
			r.SetControllerError(ctx, application, controllerName, err)
			return err
		}

		r.SetLabelsFromApplication(ctx, &digdirator, *application)
		util.SetCommonAnnotations(&digdirator)

		return nil
	})

	r.SetControllerFinishedOutcome(ctx, application, controllerName, err)

	return reconcile.Result{}, err
}
