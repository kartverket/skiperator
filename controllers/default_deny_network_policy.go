package controllers

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

//+kubebuilder:rbac:groups=core,resources=namespaces,verbs=get;list;watch
//+kubebuilder:rbac:groups=networking.k8s.io,resources=networkpolicies,verbs=get;list;watch;create;update;patch;delete

type DefaultDenyNetworkPolicyReconciler struct {
	client   client.Client
	scheme   *runtime.Scheme
	recorder record.EventRecorder
}

func (r *DefaultDenyNetworkPolicyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.client = mgr.GetClient()
	r.scheme = mgr.GetScheme()
	r.recorder = mgr.GetEventRecorderFor("default-deny-networkpolicy-controller")

	return newControllerManagedBy[*corev1.Namespace](mgr).
		For(&corev1.Namespace{}, builder.WithPredicates(
			matchesPredicate[*corev1.Namespace](isNotExcludedNamespace),
		)).
		Owns(&networkingv1.NetworkPolicy{}).
		Complete(r)
}

func (r *DefaultDenyNetworkPolicyReconciler) Reconcile(ctx context.Context, namespace *corev1.Namespace) (reconcile.Result, error) {

	instanaConfigMap := corev1.ConfigMap{}

	err := r.client.Get(ctx, types.NamespacedName{Namespace: "skiperator-system", Name: "instana-networkpolicy-config"}, &instanaConfigMap)
	if errors.IsNotFound(err) {
		r.recorder.Eventf(
			namespace,
			corev1.EventTypeWarning, "Missing",
			"Cannot find configmap named instana-networkpolicy-config in namespace skiperator-system",
		)
	} else if err != nil {
		return reconcile.Result{}, err
	}

	networkPolicy := networkingv1.NetworkPolicy{ObjectMeta: metav1.ObjectMeta{Namespace: namespace.Name, Name: "default-deny"}}
	_, err = ctrlutil.CreateOrPatch(ctx, r.client, &networkPolicy, func() error {
		// Set namespace as owner of the network policy
		err := ctrlutil.SetControllerReference(namespace, &networkPolicy, r.scheme)
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
