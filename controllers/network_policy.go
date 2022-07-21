package controllers

import (
	"context"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

//+kubebuilder:rbac:groups=skiperator.kartverket.no,resources=applications,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=networking.k8s.io,resources=networkpolicies,verbs=get;list;watch;create;update;patch;delete

type NetworkPolicyReconciler struct {
	client client.Client
	scheme *runtime.Scheme
}

func (r *NetworkPolicyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.client = mgr.GetClient()
	r.scheme = mgr.GetScheme()

	return ctrl.NewControllerManagedBy(mgr).
		For(&skiperatorv1alpha1.Application{}).
		Owns(&networkingv1.NetworkPolicy{}).
		Complete(r)
}

func (r *NetworkPolicyReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	// Fetch application and fill defaults
	application := skiperatorv1alpha1.Application{}
	err := r.client.Get(ctx, req.NamespacedName, &application)
	if err != nil {
		err = client.IgnoreNotFound(err)
		return reconcile.Result{}, err
	}
	application.FillDefaults()

	networkPolicy := networkingv1.NetworkPolicy{ObjectMeta: metav1.ObjectMeta{Namespace: req.Namespace, Name: req.Name}}
	_, err = ctrlutil.CreateOrPatch(ctx, r.client, &networkPolicy, func() error {
		// Set application as owner of the network policy
		err = ctrlutil.SetControllerReference(&application, &networkPolicy, r.scheme)
		if err != nil {
			return err
		}

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
			if isExternal(hostname) {
				external = true
			} else {
				internal = true
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
		count = 3
		count += len(application.Spec.AccessPolicy.Outbound.Rules)
		if len(application.Spec.AccessPolicy.Outbound.External) > 0 {
			count += 1
		}
		networkPolicy.Spec.Egress = make([]networkingv1.NetworkPolicyEgressRule, 3, count)

		// Egress rule for Internet
		networkPolicy.Spec.Egress[0].To = make([]networkingv1.NetworkPolicyPeer, 1)

		networkPolicy.Spec.Egress[0].To[0].IPBlock = &networkingv1.IPBlock{}
		networkPolicy.Spec.Egress[0].To[0].IPBlock.CIDR = "0.0.0.0/0"
		networkPolicy.Spec.Egress[0].To[0].IPBlock.Except = []string{"10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"}

		// Egress rule for DNS
		networkPolicy.Spec.Egress[1].To = make([]networkingv1.NetworkPolicyPeer, 1)
		networkPolicy.Spec.Egress[1].Ports = make([]networkingv1.NetworkPolicyPort, 2)

		networkPolicy.Spec.Egress[1].To[0].NamespaceSelector = &metav1.LabelSelector{}
		labels = map[string]string{"kubernetes.io/metadata.name": "kube-system"}
		networkPolicy.Spec.Egress[1].To[0].NamespaceSelector.MatchLabels = labels

		networkPolicy.Spec.Egress[1].To[0].PodSelector = &metav1.LabelSelector{}
		labels = map[string]string{"k8s-app": "kube-dns"}
		networkPolicy.Spec.Egress[1].To[0].PodSelector.MatchLabels = labels

		port := intstr.FromInt(53)
		networkPolicy.Spec.Egress[1].Ports[0].Port = &port
		protocol := new(corev1.Protocol)
		*protocol = corev1.ProtocolTCP
		networkPolicy.Spec.Egress[1].Ports[0].Protocol = protocol
		networkPolicy.Spec.Egress[1].Ports[1].Port = &port
		protocol = new(corev1.Protocol)
		*protocol = corev1.ProtocolUDP
		networkPolicy.Spec.Egress[1].Ports[1].Protocol = protocol

		// Egress rule for Istio XDS
		networkPolicy.Spec.Egress[2].To = make([]networkingv1.NetworkPolicyPeer, 1)
		networkPolicy.Spec.Egress[2].Ports = make([]networkingv1.NetworkPolicyPort, 1)

		networkPolicy.Spec.Egress[2].To[0].NamespaceSelector = &metav1.LabelSelector{}
		labels = map[string]string{"kubernetes.io/metadata.name": "istio-system"}
		networkPolicy.Spec.Egress[2].To[0].NamespaceSelector.MatchLabels = labels

		networkPolicy.Spec.Egress[2].To[0].PodSelector = &metav1.LabelSelector{}
		labels = map[string]string{"app": "istiod"}
		networkPolicy.Spec.Egress[2].To[0].PodSelector.MatchLabels = labels

		port = intstr.FromInt(15012)
		networkPolicy.Spec.Egress[2].Ports[0].Port = &port

		// Egress rules for internal peers
		for _, rule := range application.Spec.AccessPolicy.Outbound.Rules {
			if rule.Namespace == "" {
				rule.Namespace = application.Namespace
			}

			svc := corev1.Service{}
			err = r.client.Get(ctx, types.NamespacedName{Namespace: rule.Namespace, Name: rule.Application}, &svc)
			if err != nil {
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

		// Egress rule for egress gateways
		if len(application.Spec.AccessPolicy.Outbound.External) > 0 {
			// Generate list of all unique external ports
			portSet := make(map[int]struct{})
			for _, rule := range application.Spec.AccessPolicy.Outbound.External {
				for _, port := range rule.Ports {
					portSet[port.Port] = struct{}{}
				}
			}
			ports := make([]int, 0, len(portSet))
			for port := range portSet {
				ports = append(ports, port)
			}

			networkPolicy.Spec.Egress = append(networkPolicy.Spec.Egress, networkingv1.NetworkPolicyEgressRule{})
			egress := &networkPolicy.Spec.Egress[len(networkPolicy.Spec.Egress)-1]

			egress.To = make([]networkingv1.NetworkPolicyPeer, 1)
			egress.Ports = make([]networkingv1.NetworkPolicyPort, len(ports))

			egress.To[0].NamespaceSelector = &metav1.LabelSelector{}
			labels = map[string]string{"kubernetes.io/metadata.name": "istio-system"}
			egress.To[0].NamespaceSelector.MatchLabels = labels

			egress.To[0].PodSelector = &metav1.LabelSelector{}
			labels = map[string]string{"egress": "external"}
			egress.To[0].PodSelector.MatchLabels = labels

			for i, port := range ports {
				port := intstr.FromInt(port)
				egress.Ports[i].Port = &port
			}
		}

		return nil
	})
	return reconcile.Result{}, err
}
