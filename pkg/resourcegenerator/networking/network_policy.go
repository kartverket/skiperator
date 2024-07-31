package networking

import (
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/api/v1alpha1/podtypes"
	"github.com/kartverket/skiperator/pkg/util"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	GrafanaAgentName      = "grafana-agent"
	GrafanaAgentNamespace = GrafanaAgentName
)

type NetPolOpts struct {
	AccessPolicy     *podtypes.AccessPolicy
	Ingresses        *[]string
	Port             *int
	RelatedServices  *[]corev1.Service
	Namespace        string
	Namespaces       *corev1.NamespaceList
	Name             string
	PrometheusConfig *skiperatorv1alpha1.PrometheusConfig
	IstioEnabled     bool
}

func CreateNetPolSpec(opts NetPolOpts) *networkingv1.NetworkPolicySpec {
	ingressRules := getIngressRules(opts)
	egressRules := getEgressRules(opts)

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

func getEgressRules(opts NetPolOpts) []networkingv1.NetworkPolicyEgressRule {
	var egressRules []networkingv1.NetworkPolicyEgressRule
	accessPolicy := opts.AccessPolicy
	namespace := opts.Namespace
	namespaces := *opts.Namespaces
	availableServices := *opts.RelatedServices

	// Egress rules for internal peers
	if accessPolicy == nil || availableServices == nil {
		return egressRules
	}

	for _, outboundRule := range (*accessPolicy).Outbound.Rules {
		if outboundRule.Namespace == "" && outboundRule.NamespacesByLabel == nil {
			outboundRule.Namespace = namespace
		}

		relatedService, isApplicationAvailable := getRelatedService(availableServices, outboundRule, namespaces)

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
						NamespaceSelector: getNamespaceSelector(outboundRule, namespace),
					},
				},
			}

			egressRules = append(egressRules, egressRuleForOutboundRule)
		}

	}

	return egressRules
}

func getRelatedService(services []corev1.Service, rule podtypes.InternalRule, namespaces corev1.NamespaceList) (corev1.Service, bool) {
	for _, service := range services {
		if service.Name == rule.Application {

			if service.Namespace == rule.Namespace {
				return service, true
			}

			if rule.NamespacesByLabel != nil {
				if namespaceMatchesNamespacesByLabel(rule.NamespacesByLabel, namespaces) {
					return service, true
				}
			}

		}

	}

	return corev1.Service{}, false

}

func namespaceMatchesNamespacesByLabel(namespacesByLabel map[string]string, namespaces corev1.NamespaceList) bool {
	for _, namespace := range namespaces.Items {
		if namespace.Labels != nil {
			for key, value := range namespacesByLabel {
				if namespace.Labels[key] == value {
					return true
				}
			}
		}
	}

	return false
}

func getIngressRules(opts NetPolOpts) []networkingv1.NetworkPolicyIngressRule {
	var ingressRules []networkingv1.NetworkPolicyIngressRule

	if opts.Ingresses != nil && opts.Port != nil && len(*opts.Ingresses) > 0 {
		if hasInternalIngress(*opts.Ingresses) {
			ingressRules = append(ingressRules, getGatewayIngressRule(*opts.Port, true))
		}

		if hasExternalIngress(*opts.Ingresses) {
			ingressRules = append(ingressRules, getGatewayIngressRule(*opts.Port, false))
		}
	}

	// Allow grafana-agent to scrape
	if opts.IstioEnabled {
		promScrapeRuleAlloy := networkingv1.NetworkPolicyIngressRule{
			From: []networkingv1.NetworkPolicyPeer{
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
					Port: util.PointTo(util.IstioMetricsPortName),
				},
			},
		}
		
		promScrapeRule := networkingv1.NetworkPolicyIngressRule{
			From: []networkingv1.NetworkPolicyPeer{
				{
					NamespaceSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"kubernetes.io/metadata.name": GrafanaAgentNamespace},
					},
					PodSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app.kubernetes.io/instance": GrafanaAgentName,
							"app.kubernetes.io/name":     GrafanaAgentName,
						},
					},
				},
			},
			Ports: []networkingv1.NetworkPolicyPort{
				{
					Port: util.PointTo(util.IstioMetricsPortName),
				},
			},
		}


		
		ingressRules = append(ingressRules, promScrapeRule)
		ingressRules = append(ingressRules, promScrapeRuleAlloy)
	}

	if opts.AccessPolicy == nil {
		return ingressRules
	}

	if opts.AccessPolicy.Inbound != nil {
		inboundTrafficIngressRule := networkingv1.NetworkPolicyIngressRule{
			From: getInboundPolicyPeers(opts.AccessPolicy.Inbound.Rules, opts.Namespace),
			Ports: []networkingv1.NetworkPolicyPort{
				{
					Port: util.PointTo(intstr.FromInt(*opts.Port)),
				},
			},
		}

		ingressRules = append(ingressRules, inboundTrafficIngressRule)
	}

	return ingressRules
}

func getInboundPolicyPeers(inboundRules []podtypes.InternalRule, namespace string) []networkingv1.NetworkPolicyPeer {
	var policyPeers []networkingv1.NetworkPolicyPeer

	for _, inboundRule := range inboundRules {

		policyPeers = append(policyPeers, networkingv1.NetworkPolicyPeer{
			NamespaceSelector: getNamespaceSelector(inboundRule, namespace),
			PodSelector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": inboundRule.Application},
			},
		})
	}

	return policyPeers
}

func getNamespaceSelector(rule podtypes.InternalRule, namespace string) *metav1.LabelSelector {
	if rule.Namespace != "" {
		return &metav1.LabelSelector{
			MatchLabels: map[string]string{"kubernetes.io/metadata.name": rule.Namespace},
		}
	}

	if rule.NamespacesByLabel != nil {
		return &metav1.LabelSelector{
			MatchLabels: rule.NamespacesByLabel,
		}
	}

	return &metav1.LabelSelector{
		MatchLabels: map[string]string{"kubernetes.io/metadata.name": namespace},
	}
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
