package dynamic

import (
	skiperatorv1alpha1 "github.com/kartverket/skiperator/v2/api/v1alpha1"
	"github.com/kartverket/skiperator/v2/api/v1alpha1/podtypes"
	"github.com/kartverket/skiperator/v2/pkg/reconciliation"
	"github.com/kartverket/skiperator/v2/pkg/util"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"net"
	"slices"
	"strings"
)

func init() {
	multiGenerator.Register(reconciliation.ApplicationType, generateForCommon)
	multiGenerator.Register(reconciliation.JobType, generateForCommon)
}

// TODO fix mess
func generateForCommon(r reconciliation.Reconciliation) error {
	ctxLog := r.GetLogger()
	ctxLog.Debug("Attempting to generate network policy for skipobj", "skipobj", r.GetSKIPObject().GetName())

	object := r.GetSKIPObject()
	name := object.GetName()
	namespace := object.GetNamespace()
	if r.GetType() == reconciliation.JobType {
		name = util.ResourceNameWithKindPostfix(name, object.GetObjectKind().GroupVersionKind().Kind)
	}

	networkPolicy := networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
	}

	accessPolicy := object.GetCommonSpec().AccessPolicy
	var ingresses []string
	var inboundPort int32
	if r.GetType() == reconciliation.ApplicationType {
		ingresses = object.(*skiperatorv1alpha1.Application).Spec.Ingresses
		inboundPort = int32(object.(*skiperatorv1alpha1.Application).Spec.Port)
	}

	ingressRules := getIngressRules(accessPolicy, ingresses, r.IsIstioEnabled(), namespace, inboundPort)
	egressRules := getEgressRules(accessPolicy, object)

	netpolSpec := networkingv1.NetworkPolicySpec{
		PodSelector: metav1.LabelSelector{MatchLabels: util.GetPodAppSelector(name)},
		Ingress:     ingressRules,
		Egress:      egressRules,
		PolicyTypes: getPolicyTypes(ingressRules, egressRules),
	}

	if len(ingressRules) == 0 && len(egressRules) == 0 {
		ctxLog.Debug("No rules for networkpolicy, skipping", "type", r.GetType(), "namespace", namespace)
		return nil
	}

	networkPolicy.Spec = netpolSpec
	r.AddResource(&networkPolicy)
	ctxLog.Debug("Finished generating networkpolicy", "type", r.GetType(), "namespace", namespace)
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

func getCloudSQLEgressRule(skipObject skiperatorv1alpha1.SKIPObject) networkingv1.NetworkPolicyEgressRule {
	return networkingv1.NetworkPolicyEgressRule{
		To: []networkingv1.NetworkPolicyPeer{
			{
				IPBlock: &networkingv1.IPBlock{
					CIDR: skipObject.GetCommonSpec().GCP.CloudSQLProxy.IP + "/32",
				},
			},
		},
		Ports: []networkingv1.NetworkPolicyPort{{
			Port:     util.PointTo(intstr.FromInt(3307)),
			Protocol: util.PointTo(v1.ProtocolTCP),
		}},
	}
}

func getEgressRules(accessPolicy *podtypes.AccessPolicy, skipObject skiperatorv1alpha1.SKIPObject) []networkingv1.NetworkPolicyEgressRule {
	var egressRules []networkingv1.NetworkPolicyEgressRule

	if util.IsCloudSqlProxyEnabled(skipObject.GetCommonSpec().GCP) {
		egressRules = append(egressRules, getCloudSQLEgressRule(skipObject))
	}

	if accessPolicy == nil || accessPolicy.Outbound == nil {
		return egressRules
	}

	for _, rule := range accessPolicy.Outbound.Rules {
		if rule.Ports == nil {
			continue
		}
		egressRules = append(egressRules, getEgressRule(rule, skipObject.GetNamespace()))
	}

	for _, externalRule := range accessPolicy.Outbound.External {
		if externalRule.Ports == nil || externalRule.Ip == "" || net.ParseIP(externalRule.Ip) == nil {
			continue
		}
		egressRules = append(egressRules, getIPExternalRule(externalRule))
	}

	return egressRules
}

func getIPExternalRule(externalRule podtypes.ExternalRule) networkingv1.NetworkPolicyEgressRule {
	externalRuleForIP := networkingv1.NetworkPolicyEgressRule{
		To: []networkingv1.NetworkPolicyPeer{
			{
				IPBlock: &networkingv1.IPBlock{
					CIDR: externalRule.Ip + "/32",
				},
			},
		},
		Ports: mapExternalPortsToNetworkPolicyPorts(externalRule.Ports),
	}
	return externalRuleForIP
}

func mapExternalPortsToNetworkPolicyPorts(externalPorts []podtypes.ExternalPort) []networkingv1.NetworkPolicyPort {
	var ports []networkingv1.NetworkPolicyPort
	for _, externalPort := range externalPorts {
		ports = append(ports, networkingv1.NetworkPolicyPort{
			Port:     util.PointTo(intstr.FromInt(externalPort.Port)),
			Protocol: util.PointTo(v1.ProtocolTCP),
		})
	}
	return ports
}

func getEgressRule(outboundRule podtypes.InternalRule, namespace string) networkingv1.NetworkPolicyEgressRule {
	slices.SortFunc(outboundRule.Ports, sortNetPolPorts)
	egressRuleForOutboundRule := networkingv1.NetworkPolicyEgressRule{
		To: []networkingv1.NetworkPolicyPeer{
			{
				PodSelector: &metav1.LabelSelector{
					MatchLabels: util.GetPodAppSelector(outboundRule.Application),
				},
				NamespaceSelector: getNamespaceSelector(outboundRule, namespace),
			},
		},
		Ports: outboundRule.Ports,
	}
	return egressRuleForOutboundRule
}

// TODO Clean up better
func getIngressRules(accessPolicy *podtypes.AccessPolicy, ingresses []string, istioEnabled bool, namespace string, port int32) []networkingv1.NetworkPolicyIngressRule {
	var ingressRules []networkingv1.NetworkPolicyIngressRule

	if ingresses != nil && len(ingresses) > 0 {
		if hasInternalIngress(ingresses) {
			ingressRules = append(ingressRules, getGatewayIngressRule(true, port))
		}

		if hasExternalIngress(ingresses) {
			ingressRules = append(ingressRules, getGatewayIngressRule(false, port))
		}
	}

	// Allow grafana-alloy to scrape
	if istioEnabled {
		promScrapeRuleAlloy := networkingv1.NetworkPolicyIngressRule{
			From: []networkingv1.NetworkPolicyPeer{
				{
					NamespaceSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"kubernetes.io/metadata.name": AlloyAgentNamespace},
					},
					PodSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app.kubernetes.io/instance": AlloyAgentName,
							"app.kubernetes.io/name":     AlloyAgentName,
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

		ingressRules = append(ingressRules, promScrapeRuleAlloy)
	}

	if accessPolicy == nil {
		return ingressRules
	}

	if accessPolicy.Inbound != nil {
		inboundTrafficIngressRule := networkingv1.NetworkPolicyIngressRule{
			From: getInboundPolicyPeers(accessPolicy.Inbound.Rules, namespace),
		}
		if port != 0 {
			inboundTrafficIngressRule.Ports = []networkingv1.NetworkPolicyPort{{Port: util.PointTo(intstr.FromInt32(port))}}
		}
		ingressRules = append(ingressRules, inboundTrafficIngressRule)
	}

	return ingressRules
}

// TODO investigate if we can just return nil if SKIPJob
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

func getNamespaceSelector(rule podtypes.InternalRule, appNamespace string) *metav1.LabelSelector {
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
		MatchLabels: map[string]string{"kubernetes.io/metadata.name": appNamespace},
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

func getGatewayIngressRule(isInternal bool, port int32) networkingv1.NetworkPolicyIngressRule {
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
	}
	if port != 0 {
		ingressRule.Ports = []networkingv1.NetworkPolicyPort{{Port: util.PointTo(intstr.FromInt32(port))}}
	}
	return ingressRule
}

// TODO Should be in constants or something
func getIngressGatewayLabel(isInternal bool) map[string]string {
	if isInternal {
		return map[string]string{"app": "istio-ingress-internal"}
	} else {
		return map[string]string{"app": "istio-ingress-external"}
	}
}

var sortNetPolPorts = func(a networkingv1.NetworkPolicyPort, b networkingv1.NetworkPolicyPort) int {
	switch {
	case a.Port.Type != b.Port.Type:
		// different types, can't compare
		return 0
	case a.Port.Type == intstr.String && b.Port.Type == intstr.String:
		// lexicographical order
		return strings.Compare(a.Port.StrVal, b.Port.StrVal)
	case a.Port.IntValue() < b.Port.IntValue():
		return -1
	case a.Port.IntValue() > b.Port.IntValue():
		return 1
	default:
		// we should never be here ¯\_(ツ)_/¯
		return 0
	}
}
