package dynamic

import (
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/api/v1alpha1/podtypes"
	"github.com/kartverket/skiperator/pkg/reconciliation"
	"github.com/kartverket/skiperator/pkg/util"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// TODO fix mess
func generateForCommon(r reconciliation.Reconciliation) error {
	ctxLog := r.GetLogger()
	ctxLog.Debug("Attempting to generate network policy for application", "application", r.GetSKIPObject().GetName())

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

	accessPolicy := r.GetCommonSpec().AccessPolicy
	var ingresses []string
	if r.GetType() == reconciliation.ApplicationType {
		ingresses = object.(*skiperatorv1alpha1.Application).Spec.Ingresses
	}
	ingressRules := getIngressRules(accessPolicy, ingresses, r.IsIstioEnabled(), namespace)
	egressRules := getEgressRules(accessPolicy, namespace)

	netpolSpec := networkingv1.NetworkPolicySpec{
		PodSelector: metav1.LabelSelector{MatchLabels: util.GetPodAppSelector(name)},
		Ingress:     ingressRules,
		Egress:      egressRules,
		PolicyTypes: getPolicyTypes(ingressRules, egressRules),
	}

	networkPolicy.Spec = netpolSpec
	var obj client.Object = &networkPolicy
	r.AddResource(&obj)
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

func getEgressRules(accessPolicy *podtypes.AccessPolicy, appNamespace string) []networkingv1.NetworkPolicyEgressRule {
	var egressRules []networkingv1.NetworkPolicyEgressRule

	if accessPolicy == nil {
		return egressRules
	}

	for _, rule := range accessPolicy.Outbound.Rules {
		egressRules = append(egressRules, getEgressRule(rule, appNamespace))
	}

	return egressRules
}

func getEgressRule(outboundRule podtypes.InternalRule, namespace string) networkingv1.NetworkPolicyEgressRule {
	egressRuleForOutboundRule := networkingv1.NetworkPolicyEgressRule{
		To: []networkingv1.NetworkPolicyPeer{
			{
				PodSelector: &metav1.LabelSelector{
					MatchLabels: util.GetPodAppSelector(outboundRule.Application),
				},
				NamespaceSelector: getNamespaceSelector(outboundRule, namespace),
			},
		},
	}
	return egressRuleForOutboundRule
}

// TODO Clean up better
func getIngressRules(accessPolicy *podtypes.AccessPolicy, ingresses []string, istioEnabled bool, namespace string) []networkingv1.NetworkPolicyIngressRule {
	var ingressRules []networkingv1.NetworkPolicyIngressRule

	if ingresses != nil && len(ingresses) > 0 {
		if hasInternalIngress(ingresses) {
			ingressRules = append(ingressRules, getGatewayIngressRule(true))
		}

		if hasExternalIngress(ingresses) {
			ingressRules = append(ingressRules, getGatewayIngressRule(false))
		}
	}

	// Allow grafana-agent to scrape
	if istioEnabled {
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
	}

	if accessPolicy == nil {
		return ingressRules
	}

	if accessPolicy.Inbound != nil {
		inboundTrafficIngressRule := networkingv1.NetworkPolicyIngressRule{
			From: getInboundPolicyPeers(accessPolicy.Inbound.Rules, namespace),
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

func getGatewayIngressRule(isInternal bool) networkingv1.NetworkPolicyIngressRule {
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
