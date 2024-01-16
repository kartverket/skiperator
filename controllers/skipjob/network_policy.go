package skipjobcontroller

import (
	"context"

	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/networking"
	"github.com/kartverket/skiperator/pkg/util"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *SKIPJobReconciler) reconcileNetworkPolicy(ctx context.Context, skipJob *skiperatorv1alpha1.SKIPJob) (reconcile.Result, error) {
	egressServices, err := r.GetEgressServices(ctx, skipJob, skipJob.Spec.Container.AccessPolicy)
	if err != nil {
		return util.RequeueWithError(err)
	}

	namespaces, err := r.GetNamespaces(ctx, skipJob)
	if err != nil {
		return util.RequeueWithError(err)
	}

	netpolOpts := networking.NetPolOpts{
		AccessPolicy:    skipJob.Spec.Container.AccessPolicy,
		Namespace:       skipJob.Namespace,
		Namespaces:      &namespaces,
		Name:            skipJob.KindPostFixedName(),
		RelatedServices: &egressServices,
	}

	netpolSpec := networking.CreateNetPolSpec(netpolOpts)

	if netpolSpec == nil {
		return util.DoNotRequeue()
	}

	networkPolicy := networkingv1.NetworkPolicy{ObjectMeta: metav1.ObjectMeta{Namespace: skipJob.Namespace, Name: skipJob.KindPostFixedName()}}
	_, err = ctrlutil.CreateOrPatch(ctx, r.GetClient(), &networkPolicy, func() error {
		err := ctrlutil.SetControllerReference(skipJob, &networkPolicy, r.GetScheme())
		if err != nil {
			return err
		}

		util.SetCommonAnnotations(&networkPolicy)

		networkPolicy.Spec = *netpolSpec

		return nil
	})

	return util.RequeueWithError(err)
}
