package applicationcontroller

import (
	"context"

	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/util"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	GrafanaAgentName      = "grafana-agent"
	GrafanaAgentNamespace = GrafanaAgentName
)

// This is a bit hacky, but seems like best solution
func (r *ApplicationReconciler) NetworkPoliciesFromService(ctx context.Context, obj client.Object) []reconcile.Request {
	svc := obj.(*corev1.Service)

	applications := &skiperatorv1alpha1.ApplicationList{}
	err := r.GetClient().List(ctx, applications)
	if err != nil {
		return nil
	}

	requests := make([]reconcile.Request, 0, len(applications.Items))
	for _, application := range applications.Items {
		if application.Spec.AccessPolicy == nil {
			continue
		}
		for _, rule := range application.Spec.AccessPolicy.Outbound.Rules {
			if rule.Namespace == svc.Namespace && rule.Application == svc.Name {
				requests = append(requests, reconcile.Request{
					NamespacedName: types.NamespacedName{
						Namespace: application.Namespace,
						Name:      application.Name,
					},
				})
				break
			}
		}
	}
	return requests
}

func (r *ApplicationReconciler) reconcileNetworkPolicy(ctx context.Context, application *skiperatorv1alpha1.Application) (reconcile.Result, error) {
	controllerName := "NetworkPolicy"
	r.SetControllerProgressing(ctx, application, controllerName)

	networkPolicy := networkingv1.NetworkPolicy{ObjectMeta: metav1.ObjectMeta{Namespace: application.Namespace, Name: application.Name}}
	_, err := ctrlutil.CreateOrPatch(ctx, r.GetClient(), &networkPolicy, func() error {
		// Set application as owner of the network policy
		err := ctrlutil.SetControllerReference(application, &networkPolicy, r.GetScheme())
		if err != nil {
			r.SetControllerError(ctx, application, controllerName, err)
			return err
		}

		r.SetLabelsFromApplication(ctx, &networkPolicy, *application)
		util.SetCommonAnnotations(&networkPolicy)

		egressRules, err := r.getEgressRules(application, ctx)
		if err != nil {
			r.SetControllerError(ctx, application, controllerName, err)
			return err
		}

		networkPolicy.Spec = networkingv1.NetworkPolicySpec{
			PolicyTypes: []networkingv1.PolicyType{
				networkingv1.PolicyTypeIngress,
				networkingv1.PolicyTypeEgress,
			},
			PodSelector: metav1.LabelSelector{
				MatchLabels: util.GetApplicationSelector(application.Name),
			},
			Ingress: getIngressRules(application),
			Egress:  egressRules,
		}

		return nil
	})

	r.SetControllerFinishedOutcome(ctx, application, controllerName, err)

	return reconcile.Result{}, err
}

func (r ApplicationReconciler) getEgressRules(application *skiperatorv1alpha1.Application, ctx context.Context) ([]networkingv1.NetworkPolicyEgressRule, error) {
	egressRules := []networkingv1.NetworkPolicyEgressRule{}

	if application.Spec.AccessPolicy == nil {
		return egressRules, nil
	}
	// Egress rules for internal peers
	for _, outboundRule := range application.Spec.AccessPolicy.Outbound.Rules {
		if outboundRule.Namespace == "" {
			outboundRule.Namespace = application.Namespace
		}

		service := corev1.Service{}
		err := r.GetClient().Get(ctx, types.NamespacedName{Namespace: outboundRule.Namespace, Name: outboundRule.Application}, &service)
		if errors.IsNotFound(err) {
			r.GetRecorder().Eventf(
				application,
				corev1.EventTypeWarning, "Missing",
				"Cannot find application named %s in namespace %s. Egress rule will not be added.",
				outboundRule.Application, outboundRule.Namespace,
			)
			continue
		} else if err != nil {
			return egressRules, err
		}

		servicePorts := []networkingv1.NetworkPolicyPort{}
		for _, port := range service.Spec.Ports {
			servicePorts = append(servicePorts, networkingv1.NetworkPolicyPort{
				Port: util.PointTo(intstr.FromInt(int(port.Port))),
			})
		}

		egressRuleForOutboundRule := networkingv1.NetworkPolicyEgressRule{
			Ports: servicePorts,
			To: []networkingv1.NetworkPolicyPeer{
				{
					PodSelector: &metav1.LabelSelector{
						MatchLabels: service.Spec.Selector,
					},
					NamespaceSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"kubernetes.io/metadata.name": outboundRule.Namespace},
					},
				},
			},
		}

		egressRules = append(egressRules, egressRuleForOutboundRule)
	}

	return egressRules, nil
}

func getIngressRules(application *skiperatorv1alpha1.Application) []networkingv1.NetworkPolicyIngressRule {
	ingressRules := []networkingv1.NetworkPolicyIngressRule{}

	if len(application.Spec.Ingresses) > 0 {
		if hasInternalIngress(application.Spec.Ingresses) {
			ingressRules = append(ingressRules, getGatewayIngressRule(*application, true))
		}

		if hasExternalIngress(application.Spec.Ingresses) {
			ingressRules = append(ingressRules, getGatewayIngressRule(*application, false))
		}
	}

	if application.Spec.AccessPolicy == nil {
		return ingressRules
	}

	if len(application.Spec.AccessPolicy.Inbound.Rules) > 0 {
		inboundTrafficIngressRule := networkingv1.NetworkPolicyIngressRule{
			From: getInboundPolicyPeers(application),
			Ports: []networkingv1.NetworkPolicyPort{
				{
					Port: util.PointTo(intstr.FromInt(application.Spec.Port)),
				},
			},
		}

		ingressRules = append(ingressRules, inboundTrafficIngressRule)
	}

	// If Prometheus metrics are exposed, allow grafana-agent to scrape
	if application.Spec.Prometheus != nil {
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
					Port: util.PointTo(application.Spec.Prometheus.Port),
				},
			},
		}

		ingressRules = append(ingressRules, promScrapeRule)
	}

	return ingressRules
}

func getInboundPolicyPeers(application *skiperatorv1alpha1.Application) []networkingv1.NetworkPolicyPeer {
	policyPeers := []networkingv1.NetworkPolicyPeer{}

	for _, inboundRule := range application.Spec.AccessPolicy.Inbound.Rules {
		if inboundRule.Namespace == "" {
			inboundRule.Namespace = application.Namespace
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

func hasExternalIngress(applicationIngresses []string) bool {
	for _, hostname := range applicationIngresses {
		if !util.IsInternal(hostname) {
			return true
		}
	}

	return false
}

func hasInternalIngress(applicationIngresses []string) bool {
	for _, hostname := range applicationIngresses {
		if util.IsInternal(hostname) {
			return true
		}
	}

	return false
}

func getGatewayIngressRule(application skiperatorv1alpha1.Application, isInternal bool) networkingv1.NetworkPolicyIngressRule {
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
				Port: util.PointTo(intstr.FromInt(application.Spec.Port)),
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
