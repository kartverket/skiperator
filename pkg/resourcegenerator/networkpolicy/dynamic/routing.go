package dynamic

import (
	"fmt"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/reconciliation"
	"github.com/kartverket/skiperator/pkg/util"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func generateForRouting(r reconciliation.Reconciliation) error {
	ctxLog := r.GetLogger()
	ctxLog.Debug("Attempting to generate network policy for routing", "routing", r.GetReconciliationObject().GetName())
	routing, ok := r.GetReconciliationObject().(*skiperatorv1alpha1.Routing)
	if !ok {
		return fmt.Errorf("failed to cast object to Routing")
	}

	uniqueTargetApps := make(map[string]skiperatorv1alpha1.Route)
	for _, route := range routing.Spec.Routes {
		uniqueTargetApps[getNetworkPolicyName(routing, route.TargetApp)] = route
	}

	for netpolName, route := range uniqueTargetApps {
		networkPolicy := networkingv1.NetworkPolicy{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: routing.Namespace,
				Name:      netpolName,
			},
		}
		networkPolicy.Spec = networkingv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{
				MatchLabels: util.GetPodAppSelector(route.TargetApp),
			},
			PolicyTypes: []networkingv1.PolicyType{
				networkingv1.PolicyTypeIngress,
			},
			Ingress: []networkingv1.NetworkPolicyIngressRule{
				{
					From: []networkingv1.NetworkPolicyPeer{
						{
							NamespaceSelector: &metav1.LabelSelector{
								MatchLabels: util.GetIstioGatewaySelector(),
							},
							PodSelector: &metav1.LabelSelector{
								MatchLabels: util.GetIstioGatewayLabelSelector(routing.Spec.Hostname),
							},
						},
					},
					Ports: []networkingv1.NetworkPolicyPort{
						{
							Port: util.PointTo(intstr.FromInt32(route.Port)),
						},
					},
				},
			},
		}

		var obj client.Object = &networkPolicy
		r.AddResource(&obj)
	}
	ctxLog.Debug("Finished generating networkpolicy for routing", "routing", routing.Name)
	return nil
}

func getNetworkPolicyName(routing *skiperatorv1alpha1.Routing, targetApp string) string {
	return fmt.Sprintf("%s-%s-istio-ingress", routing.Name, targetApp)
}
