package dynamic

import (
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/api/v1alpha1/podtypes"
	"github.com/kartverket/skiperator/pkg/reconciliation"
	"github.com/kartverket/skiperator/pkg/util"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
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
	var hosts = skiperatorv1alpha1.NewCollection()
	var inboundPort int32

	if r.GetType() == reconciliation.ApplicationType {
		var err error
		hosts, err = object.(*skiperatorv1alpha1.Application).Spec.Hosts()
		if err != nil {
			ctxLog.Error(err, "Failed to get hosts for networkpolicy", "type", r.GetType(), "namespace", namespace)
			return err
		}
		inboundPort = int32(object.(*skiperatorv1alpha1.Application).Spec.Port)
	}

	ingressRules := getIngressRules(accessPolicy, hosts, r.IsIstioEnabled(), namespace, inboundPort)
	egressRules := getEgressRules(accessPolicy, namespace)

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

func getEgressRules(accessPolicy *podtypes.AccessPolicy, appNamespace string) []networkingv1.NetworkPolicyEgressRule {
	var egressRules []networkingv1.NetworkPolicyEgressRule

	if accessPolicy == nil || accessPolicy.Outbound == nil {
		return egressRules
	}

	for _, rule := range accessPolicy.Outbound.Rules {
		if rule.Ports == nil {
			continue
		}
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
		Ports: outboundRule.Ports,
	}
	return egressRuleForOutboundRule
}

// TODO Clean up better
func getIngressRules(accessPolicy *podtypes.AccessPolicy, hostCollection skiperatorv1alpha1.HostCollection, istioEnabled bool, namespace string, port int32) []networkingv1.NetworkPolicyIngressRule {
	var ingressRules []networkingv1.NetworkPolicyIngressRule

	for _, host := range hostCollection.AllHosts() {
		ingressRules = append(ingressRules, getGatewayIngressRule(host.Internal, port))
	}

	// Allow grafana-agent and Alloy to scrape
	if istioEnabled {
		promScrapeRuleGrafana := networkingv1.NetworkPolicyIngressRule{
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

		ingressRules = append(ingressRules, promScrapeRuleGrafana)
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
