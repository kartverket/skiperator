package namespacecontroller

import (
	"context"

	"github.com/kartverket/skiperator/pkg/util"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *NamespaceReconciler) reconcileDefaultDenyNetworkPolicy(ctx context.Context, namespace *corev1.Namespace) (reconcile.Result, error) {

	networkPolicy := networkingv1.NetworkPolicy{ObjectMeta: metav1.ObjectMeta{Namespace: namespace.Name, Name: "default-deny"}}
	_, err := ctrlutil.CreateOrPatch(ctx, r.GetClient(), &networkPolicy, func() error {
		// Set namespace as owner of the network policy
		err := ctrlutil.SetControllerReference(namespace, &networkPolicy, r.GetScheme())
		if err != nil {
			return err
		}

		networkPolicy.Spec = networkingv1.NetworkPolicySpec{
			PolicyTypes: []networkingv1.PolicyType{
				networkingv1.PolicyTypeIngress,
				networkingv1.PolicyTypeEgress,
			},
			Egress: []networkingv1.NetworkPolicyEgressRule{
				// Egress rule for Internet
				{
					To: []networkingv1.NetworkPolicyPeer{
						{
							IPBlock: &networkingv1.IPBlock{
								CIDR:   "0.0.0.0/0",
								Except: []string{"10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"},
							},
						},
					},
				},
				// Egress rule for DNS
				{
					To: []networkingv1.NetworkPolicyPeer{
						{
							NamespaceSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{"kubernetes.io/metadata.name": "kube-system"},
							},
							PodSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{"k8s-app": "kube-dns"},
							},
						},
					},
					Ports: []networkingv1.NetworkPolicyPort{
						// DNS Ports
						{
							Protocol: util.PointTo(corev1.ProtocolTCP),
							Port:     util.PointTo(intstr.FromInt(53)),
						},
						{
							Protocol: util.PointTo(corev1.ProtocolUDP),
							Port:     util.PointTo(intstr.FromInt(53)),
						},
					},
				},
				// Egress rule for Istio XDS
				{
					To: []networkingv1.NetworkPolicyPeer{
						{
							PodSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{"app": "istiod"},
							},
							NamespaceSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{"kubernetes.io/metadata.name": "istio-system"},
							},
						},
					},
					Ports: []networkingv1.NetworkPolicyPort{
						{
							Port: util.PointTo(intstr.FromInt(15012)),
						},
					},
				},
				// Egress rule for grafana-agent
				{
					To: []networkingv1.NetworkPolicyPeer{
						{
							NamespaceSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{"kubernetes.io/metadata.name": "grafana-agent"},
							},
							PodSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									"app.kubernetes.io/instance": "grafana-agent",
									"app.kubernetes.io/name":     "grafana-agent",
								},
							},
						},
					},
					Ports: []networkingv1.NetworkPolicyPort{
						{
							Protocol: util.PointTo(corev1.ProtocolTCP),
							Port:     util.PointTo(intstr.FromInt(4317)),
						},
						{
							Protocol: util.PointTo(corev1.ProtocolTCP),
							Port:     util.PointTo(intstr.FromInt(4318)),
						},
					},
				},
			},
		}

		return nil
	})
	return reconcile.Result{}, err
}
