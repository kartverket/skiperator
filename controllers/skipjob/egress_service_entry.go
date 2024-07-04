package skipjobcontroller

import (
	"context"
	"fmt"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/istio"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/istio/serviceentry"
	"github.com/kartverket/skiperator/pkg/util"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *SKIPJobReconciler) reconcileEgressServiceEntry(ctx context.Context, skipJob *skiperatorv1alpha1.SKIPJob) (reconcile.Result, error) {
	serviceEntries, err := istio.GetServiceEntries(skipJob.Spec.Container.AccessPolicy, skipJob)
	if err != nil {
		r.EmitWarningEvent(skipJob, "ServiceEntryError", fmt.Sprintf("something went wrong when fetching service entries: %v", err.Error()))
		return util.RequeueWithError(err)
	}

	for _, serviceEntry := range serviceEntries {
		// CreateOrPatch gets the object (from cache) before the mutating function is run, masquerading actual changes
		// Restoring the Spec from a copy within the mutating func fixes this
		desiredServiceEntry := serviceEntry.DeepCopy()
		_, err := ctrlutil.CreateOrPatch(ctx, r.GetClient(), &serviceEntry, func() error {
			serviceEntry.Spec = desiredServiceEntry.Spec
			// Set application as owner of the service entry
			err := ctrlutil.SetControllerReference(skipJob, &serviceEntry, r.GetScheme())
			if err != nil {
				return err
			}
			util.SetCommonAnnotations(&serviceEntry)

			return nil
		})

		if err != nil {
			return util.RequeueWithError(err)
		}
	}

	serviceEntriesInNamespace := networkingv1beta1.ServiceEntryList{}
	err = r.GetClient().List(ctx, &serviceEntriesInNamespace, client.InNamespace(skipJob.Namespace))
	if err != nil {
		return util.RequeueWithError(err)
	}

	serviceEntriesToDelete := serviceentry.GetServiceEntriesToDelete(serviceEntriesInNamespace.Items, skipJob.Name, serviceEntries)
	for _, serviceEntry := range serviceEntriesToDelete {
		err = r.DeleteObjectIfExists(ctx, &serviceEntry)
		if err != nil {
			return util.RequeueWithError(err)
		}
	}

	return util.RequeueWithError(err)
}
