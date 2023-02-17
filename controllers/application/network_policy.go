package applicationcontroller

import (
	"context"

	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	util "github.com/kartverket/skiperator/pkg/util"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// This is a bit hacky, but seems like best solution
func (r *ApplicationReconciler) NetworkPoliciesFromService(obj client.Object) []reconcile.Request {
	ctx := context.TODO()
	svc := obj.(*corev1.Service)

	applications := &skiperatorv1alpha1.ApplicationList{}
	err := r.GetClient().List(ctx, applications)
	if err != nil {
		return nil
	}

	requests := make([]reconcile.Request, 0, len(applications.Items))
	for _, application := range applications.Items {
		for _, rule := range application.Spec.AccessPolicy.Outbound.Rules {
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

	networkPolicy := networkingv1.NetworkPolicy{ObjectMeta: metav1.ObjectMeta{Namespace: application.Namespace, Name: application.Name}}
	_, err := ctrlutil.CreateOrPatch(ctx, r.GetClient(), &networkPolicy, func() error {
		// Set application as owner of the network policy
		err := ctrlutil.SetControllerReference(application, &networkPolicy, r.GetScheme())
		if err != nil {
			r.SetControllerError(ctx, application, controllerName, err)
			return err
		}

		r.SetLabelsFromApplication(ctx, &networkPolicy, *application)
		networkPolicy.ObjectMeta.Annotations = util.CommonAnnotations

		labels := map[string]string{"app": application.Name}
		networkPolicy.Spec.PodSelector.MatchLabels = labels

		networkPolicy.Spec.PolicyTypes = []networkingv1.PolicyType{
			networkingv1.PolicyTypeIngress,
			networkingv1.PolicyTypeEgress,
		}

		// Ingress rules
		internal := false
		external := false
		for _, hostname := range application.Spec.Ingresses {
			if util.IsInternal(hostname) {
				internal = true
			} else {
				external = true
			}
		}

		count := 0
		if internal {
			count += 1
		}
		if external {
			count += 1
		}
		if len(application.Spec.AccessPolicy.Inbound.Rules) > 0 {
			count += 1
		}
		networkPolicy.Spec.Ingress = make([]networkingv1.NetworkPolicyIngressRule, 0, count)

		// Ingress rule for ingress gateways
		if internal {
			networkPolicy.Spec.Ingress = append(networkPolicy.Spec.Ingress, networkingv1.NetworkPolicyIngressRule{})
			ingress := &networkPolicy.Spec.Ingress[len(networkPolicy.Spec.Ingress)-1]

			ingress.From = make([]networkingv1.NetworkPolicyPeer, 1)
			ingress.Ports = make([]networkingv1.NetworkPolicyPort, 1)

			ingress.From[0].NamespaceSelector = &metav1.LabelSelector{}
			labels = map[string]string{"kubernetes.io/metadata.name": "istio-system"}
			ingress.From[0].NamespaceSelector.MatchLabels = labels

			ingress.From[0].PodSelector = &metav1.LabelSelector{}
			labels = map[string]string{"ingress": "internal"}
			ingress.From[0].PodSelector.MatchLabels = labels

			port := intstr.FromInt(application.Spec.Port)
			ingress.Ports[0].Port = &port
		}

		if external {
			networkPolicy.Spec.Ingress = append(networkPolicy.Spec.Ingress, networkingv1.NetworkPolicyIngressRule{})
			ingress := &networkPolicy.Spec.Ingress[len(networkPolicy.Spec.Ingress)-1]

			ingress.From = make([]networkingv1.NetworkPolicyPeer, 1)
			ingress.Ports = make([]networkingv1.NetworkPolicyPort, 1)

			ingress.From[0].NamespaceSelector = &metav1.LabelSelector{}
			labels = map[string]string{"kubernetes.io/metadata.name": "istio-system"}
			ingress.From[0].NamespaceSelector.MatchLabels = labels

			ingress.From[0].PodSelector = &metav1.LabelSelector{}
			labels = map[string]string{"ingress": "external"}
			ingress.From[0].PodSelector.MatchLabels = labels

			port := intstr.FromInt(application.Spec.Port)
			ingress.Ports[0].Port = &port
		}

		// Ingress rules for internal peers
		if len(application.Spec.AccessPolicy.Inbound.Rules) > 0 {
			networkPolicy.Spec.Ingress = append(networkPolicy.Spec.Ingress, networkingv1.NetworkPolicyIngressRule{})
			ingress := &networkPolicy.Spec.Ingress[len(networkPolicy.Spec.Ingress)-1]

			ingress.From = make([]networkingv1.NetworkPolicyPeer, len(application.Spec.AccessPolicy.Inbound.Rules))
			ingress.Ports = make([]networkingv1.NetworkPolicyPort, 1)

			for i, rule := range application.Spec.AccessPolicy.Inbound.Rules {
				if rule.Namespace == "" {
					rule.Namespace = application.Namespace
				}

				ingress.From[i].NamespaceSelector = &metav1.LabelSelector{}
				labels = map[string]string{"kubernetes.io/metadata.name": rule.Namespace}
				ingress.From[i].NamespaceSelector.MatchLabels = labels

				ingress.From[i].PodSelector = &metav1.LabelSelector{}
				labels = map[string]string{"app": rule.Application}
				ingress.From[i].PodSelector.MatchLabels = labels
			}

			port := intstr.FromInt(application.Spec.Port)
			ingress.Ports[0].Port = &port
		}

		// Egress rules
		count = len(application.Spec.AccessPolicy.Outbound.Rules)
		networkPolicy.Spec.Egress = make([]networkingv1.NetworkPolicyEgressRule, 0, count)

		// Egress rules for internal peers
		for _, rule := range application.Spec.AccessPolicy.Outbound.Rules {
			if rule.Namespace == "" {
				rule.Namespace = application.Namespace
			}

			svc := corev1.Service{}
			err = r.GetClient().Get(ctx, types.NamespacedName{Namespace: rule.Namespace, Name: rule.Application}, &svc)
			if errors.IsNotFound(err) {
				r.GetRecorder().Eventf(
					application,
					corev1.EventTypeWarning, "Missing",
					"Cannot find application named %s in namespace %s",
					rule.Application, rule.Namespace,
				)
				continue
			} else if err != nil {
				r.SetControllerError(ctx, application, controllerName, err)
				return err
			}

			networkPolicy.Spec.Egress = append(networkPolicy.Spec.Egress, networkingv1.NetworkPolicyEgressRule{})
			egress := &networkPolicy.Spec.Egress[len(networkPolicy.Spec.Egress)-1]

			egress.To = make([]networkingv1.NetworkPolicyPeer, 1)
			egress.Ports = make([]networkingv1.NetworkPolicyPort, len(svc.Spec.Ports))

			egress.To[0].NamespaceSelector = &metav1.LabelSelector{}
			labels = map[string]string{"kubernetes.io/metadata.name": rule.Namespace}
			egress.To[0].NamespaceSelector.MatchLabels = labels

			egress.To[0].PodSelector = &metav1.LabelSelector{}
			egress.To[0].PodSelector.MatchLabels = svc.Spec.Selector

			for i := range svc.Spec.Ports {
				port := intstr.FromInt(int(svc.Spec.Ports[i].Port))
				egress.Ports[i].Port = &port
			}
		}

		return nil
	})

	r.SetControllerFinishedOutcome(ctx, application, controllerName, err)

	return reconcile.Result{}, err
}
