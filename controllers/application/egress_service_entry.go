package applicationcontroller

import (
	"context"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/istio"
	"github.com/kartverket/skiperator/pkg/util"
	"golang.org/x/exp/slices"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
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

	// Clear out unused service entries
	serviceEntriesInNamespace := networkingv1beta1.ServiceEntryList{}
	err := r.GetClient().List(ctx, &serviceEntriesInNamespace, client.InNamespace(application.Namespace))
	if err != nil {
		r.SetControllerError(ctx, application, controllerName, err)
		return reconcile.Result{}, err
	}

	for _, serviceEntry := range serviceEntriesInNamespace.Items {
		applicationOwnerIndex := slices.IndexFunc(serviceEntry.GetOwnerReferences(), func(ownerReference metav1.OwnerReference) bool {
			return ownerReference.Name == application.Name
		})
		serviceEntryOwnedByThisApplication := applicationOwnerIndex != -1
		if !serviceEntryOwnedByThisApplication {
			continue
		}

		serviceEntryInApplicationSpecIndex := slices.IndexFunc(serviceEntries, func(inSpecEntry networkingv1beta1.ServiceEntry) bool {
			return inSpecEntry.Name == serviceEntry.Name
		})

		serviceEntryInApplicationSpec := serviceEntryInApplicationSpecIndex != -1
		if serviceEntryInApplicationSpec {
			continue
		}

		// Delete the rest
		err = r.GetClient().Delete(ctx, serviceEntry)
		err = client.IgnoreNotFound(err)
		if err != nil {
			r.SetControllerError(ctx, application, controllerName, err)
			return reconcile.Result{}, err
		}
	}

	r.SetControllerFinishedOutcome(ctx, application, controllerName, err)

	return reconcile.Result{}, err
}
