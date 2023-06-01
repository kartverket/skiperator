package applicationcontroller

import (
	"context"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/istio"
	"github.com/kartverket/skiperator/pkg/util"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *SKIPJobReconciler) reconcileEgressServiceEntry(ctx context.Context, skipJob *skiperatorv1alpha1.SKIPJob) (reconcile.Result, error) {
	serviceEntries := istio.GetServiceEntries(skipJob.Spec.Container.AccessPolicy, skipJob)

	for _, serviceEntry := range serviceEntries {
		_, err := ctrlutil.CreateOrPatch(ctx, r.GetClient(), &serviceEntry, func() error {
			// Set application as owner of the service entry
			err := ctrlutil.SetControllerReference(skipJob, &serviceEntry, r.GetScheme())
			if err != nil {
				return err
			}
			util.SetCommonAnnotations(&serviceEntry)

			return nil
		})

		if err != nil {
			return reconcile.Result{}, err
		}
	}

	err := r.DeleteUnusedEgresses(ctx, skipJob.Name, skipJob.Namespace, serviceEntries)
	if err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, err
}
