package applicationcontroller

import (
	"context"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/pod"
	"k8s.io/apimachinery/pkg/api/errors"

	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/util"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// This is a bit hacky, but seems like best solution
func (r *ApplicationReconciler) NetworkPoliciesFromService(ctx context.Context, obj client.Object) []reconcile.Request {
	svc := obj.(*corev1.Service)

	applications := &skiperatorv1alpha1.ApplicationList{}
	err := r.GetClient().List(ctx, applications)
	if err != nil {
		return nil
	}

	requests := make([]reconcile.Request, 0, len(applications.Items))
	for _, application := range applications.Items {
		if application.Spec.AccessPolicy == nil {
			continue
		}
		for _, rule := range (*application.Spec.AccessPolicy).Outbound.Rules {
			if rule.Namespace == svc.Namespace && rule.Application == svc.Name {
				requests = append(requests, reconcile.Request{
					NamespacedName: types.NamespacedName{
						Namespace: application.Namespace,
						Name:      application.Name,
					},
				})
				break
			}
		}
	}
	return requests
}

func (r *ApplicationReconciler) reconcileNetworkPolicy(ctx context.Context, application *skiperatorv1alpha1.Application) (reconcile.Result, error) {
	controllerName := "NetworkPolicy"
	r.SetControllerProgressing(ctx, application, controllerName)

	egressServices, err := r.getEgressServices(ctx, *application)
	if err != nil {
		return reconcile.Result{}, err
	}

	networkPolicy := networkingv1.NetworkPolicy{ObjectMeta: metav1.ObjectMeta{Namespace: application.Namespace, Name: application.Name}}
	_, err = ctrlutil.CreateOrPatch(ctx, r.GetClient(), &networkPolicy, func() error {
		// Set application as owner of the network policy
		err := ctrlutil.SetControllerReference(application, &networkPolicy, r.GetScheme())
		if err != nil {
			r.SetControllerError(ctx, application, controllerName, err)
			return err
		}

		r.SetLabelsFromApplication(ctx, &networkPolicy, *application)
		util.SetCommonAnnotations(&networkPolicy)

		networkPolicy.Spec = pod.CreateNetPolSpec(pod.NetPolOpts{
			AccessPolicy:    application.Spec.AccessPolicy,
			Ingresses:       &application.Spec.Ingresses,
			Port:            &application.Spec.Port,
			Namespace:       application.Namespace,
			Name:            application.Name,
			RelatedServices: &egressServices,
		})

		return nil
	})

	r.SetControllerFinishedOutcome(ctx, application, controllerName, err)

	return reconcile.Result{}, err
}

func (r *ApplicationReconciler) getEgressServices(ctx context.Context, application skiperatorv1alpha1.Application) ([]corev1.Service, error) {
	var egressServices []corev1.Service
	if application.Spec.AccessPolicy == nil {
		return egressServices, nil
	}

	for _, outboundRule := range application.Spec.AccessPolicy.Outbound.Rules {
		if outboundRule.Namespace == "" {
			outboundRule.Namespace = application.Namespace
		}

		service := corev1.Service{}

		err := r.GetClient().Get(ctx, client.ObjectKey{
			Namespace: outboundRule.Namespace,
			Name:      outboundRule.Application,
		}, &service)
		if errors.IsNotFound(err) {
			r.GetRecorder().Eventf(
				&application,
				corev1.EventTypeWarning, "Missing",
				"Cannot find application named %s in namespace %s. Egress rule will not be added.",
				outboundRule.Application, outboundRule.Namespace,
			)
			continue
		} else if err != nil {
			return egressServices, err
		}

		egressServices = append(egressServices, service)
	}

	return egressServices, nil
}
