package defaultdeny

import (
	"fmt"
	"github.com/kartverket/skiperator/pkg/reconciliation"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kartverket/skiperator/pkg/util"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func Generate(r reconciliation.Reconciliation) error {
	ctxLog := r.GetLogger()
	ctxLog.Debug("Attempting to generate default deny network policy for namespace", r.GetReconciliationObject().GetName())

	if r.GetType() != reconciliation.NamespaceType {
		return fmt.Errorf("default deny namespace only supports namespace type")
	}

	networkPolicy := networkingv1.NetworkPolicy{ObjectMeta: metav1.ObjectMeta{Namespace: r.GetReconciliationObject().GetName(), Name: "default-deny"}}

	networkPolicy.Spec = networkingv1.NetworkPolicySpec{
		PolicyTypes: []networkingv1.PolicyType{
			networkingv1.PolicyTypeIngress,
			networkingv1.PolicyTypeEgress,
		},
		Egress: []networkingv1.NetworkPolicyEgressRule{
			{
				To: []networkingv1.NetworkPolicyPeer{
					// Egress rule for parts of internal server network
					{
						IPBlock: &networkingv1.IPBlock{
							CIDR: "10.40.0.0/16",
						},
					},
					// Egress rule for Internet
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

	var obj client.Object = &networkPolicy
	r.AddResource(&obj)

	ctxLog.Debug("Finished generating default deny network policy for namespace", r.GetReconciliationObject().GetName())
	return nil
}
