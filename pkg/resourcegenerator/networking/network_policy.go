package networking

import (
	"github.com/kartverket/skiperator/api/v1alpha1/podtypes"
	"github.com/kartverket/skiperator/pkg/util"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type NetPolOpts struct {
	AccessPolicy    *podtypes.AccessPolicy
	Ingresses       *[]string
	Port            *int
	RelatedServices *[]corev1.Service
	Namespace       string
	Name            string
}

func CreateNetPolSpec(opts NetPolOpts) *networkingv1.NetworkPolicySpec {
	if opts.AccessPolicy == nil {
		return nil
	}

	ingressRules := getIngressRules(opts.AccessPolicy, opts.Ingresses, opts.Port, opts.Namespace)
	egressRules := getEgressRules(opts.AccessPolicy, opts.Namespace, *opts.RelatedServices)

	if len(ingressRules) > 0 || len(egressRules) > 0 {
		return &networkingv1.NetworkPolicySpec{
			PolicyTypes: getPolicyTypes(ingressRules, egressRules),
			PodSelector: metav1.LabelSelector{
				MatchLabels: util.GetPodAppSelector(opts.Name),
			},
			Ingress: ingressRules,
			Egress:  egressRules,
		}
	}

	return nil
}

func getPolicyTypes(ingressRules []networkingv1.NetworkPolicyIngressRule, egressRules []networkingv1.NetworkPolicyEgressRule) []networkingv1.PolicyType {
	var policyType []networkingv1.PolicyType

	if len(ingressRules) > 0 {
		policyType = append(policyType, networkingv1.PolicyTypeIngress)
	}

	if len(egressRules) > 0 {
		policyType = append(policyType, networkingv1.PolicyTypeEgress)
	}

	return policyType
}

func getEgressRules(accessPolicy *podtypes.AccessPolicy, namespace string, availableServices []corev1.Service) []networkingv1.NetworkPolicyEgressRule {
	var egressRules []networkingv1.NetworkPolicyEgressRule

	// Egress rules for internal peers
	if accessPolicy == nil || availableServices == nil {
		return egressRules
	}

	for _, outboundRule := range (*accessPolicy).Outbound.Rules {
		if outboundRule.Namespace == "" {
			outboundRule.Namespace = namespace
		}

		relatedService, isApplicationAvailable := getRelatedService(availableServices, outboundRule)

		if !isApplicationAvailable {
			continue
		} else {
			var servicePorts []networkingv1.NetworkPolicyPort

			for _, port := range relatedService.Spec.Ports {
				servicePorts = append(servicePorts, networkingv1.NetworkPolicyPort{
					Port: util.PointTo(intstr.FromInt(int(port.Port))),
				})
			}

			egressRuleForOutboundRule := networkingv1.NetworkPolicyEgressRule{
				Ports: servicePorts,
				To: []networkingv1.NetworkPolicyPeer{
					{
						PodSelector: &metav1.LabelSelector{
							MatchLabels: relatedService.Spec.Selector,
						},
						NamespaceSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{"kubernetes.io/metadata.name": outboundRule.Namespace},
						},
					},
				},
			}

			egressRules = append(egressRules, egressRuleForOutboundRule)
		}

	}

	return egressRules
}

func getRelatedService(services []corev1.Service, rule podtypes.InternalRule) (corev1.Service, bool) {
	for _, service := range services {
		if service.Name == rule.Application && service.Namespace == rule.Namespace {
			return service, true

		}
	}

	return corev1.Service{}, false
}

func getIngressRules(accessPolicy *podtypes.AccessPolicy, ingresses *[]string, port *int, namespace string) []networkingv1.NetworkPolicyIngressRule {
	var ingressRules []networkingv1.NetworkPolicyIngressRule

	if ingresses != nil && port != nil && len(*ingresses) > 0 {
		if hasInternalIngress(*ingresses) {
			ingressRules = append(ingressRules, getGatewayIngressRule(*port, true))
		}

		if hasExternalIngress(*ingresses) {
			ingressRules = append(ingressRules, getGatewayIngressRule(*port, false))
		}
	}

	if accessPolicy != nil && port != nil {
		if len((*accessPolicy).Inbound.Rules) > 0 {
			inboundTrafficIngressRule := networkingv1.NetworkPolicyIngressRule{
				From: getInboundPolicyPeers(accessPolicy.Inbound.Rules, namespace),
				Ports: []networkingv1.NetworkPolicyPort{
					{
						Port: util.PointTo(intstr.FromInt(*port)),
					},
				},
			}

			ingressRules = append(ingressRules, inboundTrafficIngressRule)
		}
	}

	return ingressRules
}

func getInboundPolicyPeers(inboundRules []podtypes.InternalRule, namespace string) []networkingv1.NetworkPolicyPeer {
	var policyPeers []networkingv1.NetworkPolicyPeer

	for _, inboundRule := range inboundRules {
		if inboundRule.Namespace == "" {
			inboundRule.Namespace = namespace
		}

		policyPeers = append(policyPeers, networkingv1.NetworkPolicyPeer{
			NamespaceSelector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"kubernetes.io/metadata.name": inboundRule.Namespace},
			},
			PodSelector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": inboundRule.Application},
			},
		})
	}

	return policyPeers
}

func hasExternalIngress(ingresses []string) bool {
	for _, hostname := range ingresses {
		if !util.IsInternal(hostname) {
			return true
		}
	}

	return false
}

func hasInternalIngress(ingresses []string) bool {
	for _, hostname := range ingresses {
		if util.IsInternal(hostname) {
			return true
		}
	}

	return false
}

func getGatewayIngressRule(port int, isInternal bool) networkingv1.NetworkPolicyIngressRule {
	ingressRule := networkingv1.NetworkPolicyIngressRule{
		From: []networkingv1.NetworkPolicyPeer{
			{
				NamespaceSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{"kubernetes.io/metadata.name": "istio-gateways"},
				},
				PodSelector: &metav1.LabelSelector{
					MatchLabels: getIngressGatewayLabel(isInternal),
				},
			},
		},
		Ports: []networkingv1.NetworkPolicyPort{
			{
				Port: util.PointTo(intstr.FromInt(port)),
			},
		},
	}

	return ingressRule
}

func getIngressGatewayLabel(isInternal bool) map[string]string {
	if isInternal {
		return map[string]string{"app": "istio-ingress-internal"}
	} else {
		return map[string]string{"app": "istio-ingress-external"}
	}
}
