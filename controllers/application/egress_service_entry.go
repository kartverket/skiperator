package applicationcontroller

import (
	"context"
	"fmt"

	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/istio"
	"github.com/kartverket/skiperator/pkg/util"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *ApplicationReconciler) reconcileEgressServiceEntry(ctx context.Context, application *skiperatorv1alpha1.Application) (reconcile.Result, error) {
	controllerName := "EgressServiceEntry"
	r.SetControllerProgressing(ctx, application, controllerName)

	serviceEntries, err := istio.GetServiceEntries(application.Spec.AccessPolicy, application)
	if err != nil {
		r.EmitWarningEvent(application, "ServiceEntryError", fmt.Sprintf("something went wrong when fetching service entries: %v", err.Error()))

		return reconcile.Result{}, err
	}

	for _, serviceEntry := range serviceEntries {
		// CreateOrPatch gets the object (from cache) before the mutating function is run, masquerading actual changes
		// Restoring the Spec from a copy within the mutating func fixes this
		desiredServiceEntry := serviceEntry.DeepCopy()

		shouldReconcile, err := r.ShouldReconcile(ctx, &serviceEntry)
		if err != nil {
			r.SetControllerFinishedOutcome(ctx, application, controllerName, err)
			return reconcile.Result{}, err
		}

		if !shouldReconcile {
			continue
		}

		_, err = ctrlutil.CreateOrPatch(ctx, r.GetClient(), &serviceEntry, func() error {
			serviceEntry.Spec = desiredServiceEntry.Spec
			// Set application as owner of the service entry
			err := ctrlutil.SetControllerReference(application, &serviceEntry, r.GetScheme())
			if err != nil {
				r.SetControllerError(ctx, application, controllerName, err)
				return err
			}
			r.SetLabelsFromApplication(&serviceEntry, *application)
			util.SetCommonAnnotations(&serviceEntry)

			return nil
		})

		if err != nil {
			r.SetControllerError(ctx, application, controllerName, err)
			return reconcile.Result{}, err
		}
	}

	serviceEntriesInNamespace := networkingv1beta1.ServiceEntryList{}
	err = r.GetClient().List(ctx, &serviceEntriesInNamespace, client.InNamespace(application.Namespace))
	if err != nil {
		r.SetControllerError(ctx, application, controllerName, err)
		return reconcile.Result{}, err
	}

	serviceEntriesToDelete := istio.GetServiceEntriesToDelete(serviceEntriesInNamespace.Items, application.Name, serviceEntries)
	for _, serviceEntry := range serviceEntriesToDelete {
		err = r.GetClient().Delete(ctx, &serviceEntry)
		err = client.IgnoreNotFound(err)
		if err != nil {
			r.SetControllerError(ctx, application, controllerName, err)
			return reconcile.Result{}, err
		}
	}

	r.SetControllerFinishedOutcome(ctx, application, controllerName, err)

	return reconcile.Result{}, err
}
