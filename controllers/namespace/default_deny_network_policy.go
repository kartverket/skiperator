package namespacecontroller

import (
	"context"

	"github.com/kartverket/skiperator/pkg/util"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *NamespaceReconciler) reconcileDefaultDenyNetworkPolicy(ctx context.Context, namespace *corev1.Namespace) (reconcile.Result, error) {
	cmapNamespacedName := types.NamespacedName{Namespace: "skiperator-system", Name: "instana-networkpolicy-config"}
	instanaConfigMap, err := util.GetConfigMap(r.GetClient(), ctx, cmapNamespacedName)

	if !util.ErrIsMissing(r.GetRecorder(), err, "Cannot find configmap named instana-networkpolicy-config in namespace skiperator-system", namespace) {
		return reconcile.Result{}, err
	}

	networkPolicy := networkingv1.NetworkPolicy{ObjectMeta: metav1.ObjectMeta{Namespace: namespace.Name, Name: "default-deny"}}
	_, err = ctrlutil.CreateOrPatch(ctx, r.GetClient(), &networkPolicy, func() error {
		// Set namespace as owner of the network policy
		err := ctrlutil.SetControllerReference(namespace, &networkPolicy, r.GetScheme())
		if err != nil {
			return err
		}

		networkPolicy.Spec.PolicyTypes = []networkingv1.PolicyType{
			networkingv1.PolicyTypeIngress,
			networkingv1.PolicyTypeEgress,
		}

		// Egress rules
		networkPolicy.Spec.Egress = make([]networkingv1.NetworkPolicyEgressRule, 4, 4)

		// Egress rule for Internet
		networkPolicy.Spec.Egress[0].To = make([]networkingv1.NetworkPolicyPeer, 1)

		networkPolicy.Spec.Egress[0].To[0].IPBlock = &networkingv1.IPBlock{}
		networkPolicy.Spec.Egress[0].To[0].IPBlock.CIDR = "0.0.0.0/0"
		networkPolicy.Spec.Egress[0].To[0].IPBlock.Except = []string{"10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"}

		// Egress rule for DNS
		networkPolicy.Spec.Egress[1].To = make([]networkingv1.NetworkPolicyPeer, 1)
		networkPolicy.Spec.Egress[1].Ports = make([]networkingv1.NetworkPolicyPort, 2)

		networkPolicy.Spec.Egress[1].To[0].NamespaceSelector = &metav1.LabelSelector{}
		labels := map[string]string{"kubernetes.io/metadata.name": "kube-system"}
		networkPolicy.Spec.Egress[1].To[0].NamespaceSelector.MatchLabels = labels

		networkPolicy.Spec.Egress[1].To[0].PodSelector = &metav1.LabelSelector{}
		labels = map[string]string{"k8s-app": "kube-dns"}
		networkPolicy.Spec.Egress[1].To[0].PodSelector.MatchLabels = labels

		dnsPort := intstr.FromInt(53)
		networkPolicy.Spec.Egress[1].Ports[0].Port = &dnsPort
		protocol := new(corev1.Protocol)
		*protocol = corev1.ProtocolTCP
		networkPolicy.Spec.Egress[1].Ports[0].Protocol = protocol
		networkPolicy.Spec.Egress[1].Ports[1].Port = &dnsPort
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

		xdsPort := intstr.FromInt(15012)
		networkPolicy.Spec.Egress[2].Ports[0].Port = &xdsPort
		// Egress rule for instana-agents
		if instanaConfigMap.Data != nil {
			networkPolicy.Spec.Egress[3].To = make([]networkingv1.NetworkPolicyPeer, 1)
			networkPolicy.Spec.Egress[3].To[0].IPBlock = &networkingv1.IPBlock{}
			networkPolicy.Spec.Egress[3].To[0].IPBlock.CIDR = instanaConfigMap.Data["cidrBlock"]
		}

		return nil
	})
	return reconcile.Result{}, err
}
