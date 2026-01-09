package defaultdeny

import (
	"fmt"

	"github.com/kartverket/skiperator/v2/internal/config"
	"github.com/kartverket/skiperator/v2/pkg/reconciliation"
	"github.com/kartverket/skiperator/v2/pkg/util"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type DefaultDenyNetworkPolicy struct {
	SKIPClusterList  *config.SKIPClusterList
	exclusionEnabled bool
}

func NewDefaultDenyNetworkPolicy(clusters *config.SKIPClusterList, exclusionEnabled bool) (*DefaultDenyNetworkPolicy, error) {
	if clusters == nil && exclusionEnabled {
		return nil, fmt.Errorf("unable to create default deny network policy: SKIPClusterList is nil")
	}
	return &DefaultDenyNetworkPolicy{
		SKIPClusterList:  clusters,
		exclusionEnabled: exclusionEnabled,
	}, nil
}

func (ddnp *DefaultDenyNetworkPolicy) Generate(r reconciliation.Reconciliation) error {
	ctxLog := r.GetLogger()
	ctxLog.Debug("Attempting to generate default deny network policy for namespace", "namespace", r.GetSKIPObject().GetName())

	if r.GetType() != reconciliation.NamespaceType {
		return fmt.Errorf("default deny namespace only supports namespace type")
	}

	networkPolicy := networkingv1.NetworkPolicy{ObjectMeta: metav1.ObjectMeta{Namespace: r.GetSKIPObject().GetName(), Name: "default-deny"}}

	ipBlock := &networkingv1.IPBlock{
		CIDR: "10.40.0.0/16",
	}

	if ddnp.exclusionEnabled {
		ipBlock.Except = ddnp.SKIPClusterList.CombinedCIDRS()
	}

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
						IPBlock: ipBlock,
					},
					// Egress rule for internal load balancer on atgcp1-sandbox
					{
						IPBlock: &networkingv1.IPBlock{
							CIDR: "10.142.5.0/28",
						},
					},
					// Egress rule for internal load balancer on atgcp1-dev
					{
						IPBlock: &networkingv1.IPBlock{
							CIDR: "10.142.3.0/28",
						},
					},
					// Egress rule for internal load balancer on atgcp1-prod
					{
						IPBlock: &networkingv1.IPBlock{
							CIDR: "10.142.1.0/28",
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
						Port:     util.PointTo(intstr.FromInt32(53)),
					},
					{
						Protocol: util.PointTo(corev1.ProtocolUDP),
						Port:     util.PointTo(intstr.FromInt32(53)),
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
						Port: util.PointTo(intstr.FromInt32(15012)),
					},
				},
			},
			// Egress rule for grafana-alloy
			{
				To: []networkingv1.NetworkPolicyPeer{
					{
						NamespaceSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{"kubernetes.io/metadata.name": "grafana-alloy"},
						},
						PodSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"app.kubernetes.io/instance": "alloy",
								"app.kubernetes.io/name":     "alloy",
							},
						},
					},
				},
				Ports: []networkingv1.NetworkPolicyPort{
					{
						Protocol: util.PointTo(corev1.ProtocolTCP),
						Port:     util.PointTo(intstr.FromInt32(4317)),
					},
					{
						Protocol: util.PointTo(corev1.ProtocolTCP),
						Port:     util.PointTo(intstr.FromInt32(4318)),
					},
				},
			},
		},
	}

	r.AddResource(&networkPolicy)

	ctxLog.Debug("Finished generating default deny network policy for namespace", "namespace", r.GetSKIPObject().GetName())
	return nil
}
