package applicationcontroller

import (
	"context"

	"github.com/kartverket/skiperator/pkg/resourcegenerator/networking"
	"sigs.k8s.io/controller-runtime/pkg/client"

	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/util"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *ApplicationReconciler) reconcileNetworkPolicy(ctx context.Context, application *skiperatorv1alpha1.Application) (reconcile.Result, error) {
	controllerName := "NetworkPolicy"
	r.SetControllerProgressing(ctx, application, controllerName)

	egressServices, err := r.GetEgressServices(ctx, application, application.Spec.AccessPolicy)
	if err != nil {
		r.SetControllerFinishedOutcome(ctx, application, controllerName, err)
		return util.RequeueWithError(err)
	}

	namespaces, err := r.GetNamespaces(ctx, application)
	if err != nil {
		r.SetControllerFinishedOutcome(ctx, application, controllerName, err)
		return util.RequeueWithError(err)
	}

	networkPolicy := networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: application.Namespace,
			Name:      application.Name,
		},
	}

	shouldReconcile, err := r.ShouldReconcile(ctx, &networkPolicy)
	if err != nil || !shouldReconcile {
		r.SetControllerFinishedOutcome(ctx, application, controllerName, err)
		return util.RequeueWithError(err)
	}

	netpolSpec := networking.CreateNetPolSpec(
		networking.NetPolOpts{
			AccessPolicy:     application.Spec.AccessPolicy,
			Ingresses:        &application.Spec.Ingresses,
			Port:             &application.Spec.Port,
			Namespace:        application.Namespace,
			Namespaces:       &namespaces,
			Name:             application.Name,
			RelatedServices:  &egressServices,
			PrometheusConfig: application.Spec.Prometheus,
			IstioEnabled:     r.IsIstioEnabledForNamespace(ctx, application.Namespace),
		},
	)

	if netpolSpec == nil {
		err = client.IgnoreNotFound(r.GetClient().Delete(ctx, &networkPolicy))
		r.SetControllerFinishedOutcome(ctx, application, controllerName, err)
		return util.RequeueWithError(err)
	}

	_, err = ctrlutil.CreateOrPatch(ctx, r.GetClient(), &networkPolicy, func() error {
		// Set application as owner of the network policy
		err := ctrlutil.SetControllerReference(application, &networkPolicy, r.GetScheme())
		if err != nil {
			r.SetControllerError(ctx, application, controllerName, err)
			return err
		}

		r.SetLabelsFromApplication(&networkPolicy, *application)
		util.SetCommonAnnotations(&networkPolicy)

		networkPolicy.Spec = *netpolSpec

		return nil
	})

	r.SetControllerFinishedOutcome(ctx, application, controllerName, err)

	return util.RequeueWithError(err)
}
