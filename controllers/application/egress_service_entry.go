package applicationcontroller

import (
	"context"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/istio"
	"github.com/kartverket/skiperator/pkg/util"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *ApplicationReconciler) reconcileEgressServiceEntry(ctx context.Context, application *skiperatorv1alpha1.Application) (reconcile.Result, error) {
	controllerName := "EgressServiceEntry"
	r.SetControllerProgressing(ctx, application, controllerName)

	serviceEntries := istio.GetServiceEntries(application.Spec.AccessPolicy, application)

	for _, serviceEntry := range serviceEntries {
		_, err := ctrlutil.CreateOrPatch(ctx, r.GetClient(), &serviceEntry, func() error {
			// Set application as owner of the service entry
			err := ctrlutil.SetControllerReference(application, &serviceEntry, r.GetScheme())
			if err != nil {
				r.SetControllerError(ctx, application, controllerName, err)
				return err
			}
			r.SetLabelsFromApplication(ctx, &serviceEntry, *application)
			util.SetCommonAnnotations(&serviceEntry)

			return nil
		})

		if err != nil {
			r.SetControllerError(ctx, application, controllerName, err)
			return reconcile.Result{}, err
		}
	}

	err := r.DeleteUnusedEgresses(ctx, application.Name, application.Namespace, serviceEntries)
	if err != nil {
		r.SetControllerError(ctx, application, controllerName, err)
		return reconcile.Result{}, err
	}

	r.SetControllerFinishedOutcome(ctx, application, controllerName, err)

	return reconcile.Result{}, err
}
