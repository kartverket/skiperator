package routingcontroller

import (
	"context"
	"fmt"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/util"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"slices"
)

func (r *RoutingReconciler) reconcileNetworkPolicy(ctx context.Context, routing *skiperatorv1alpha1.Routing) (reconcile.Result, error) {
	var err error

	// Get map of unique network policies: map[networkPolicyName]targetApp
	uniqueTargetApps := make(map[string]string)
	for _, route := range routing.Spec.Routes {
		uniqueTargetApps[getNetworkPolicyName(routing, route.TargetApp)] = route.TargetApp
	}

	for netpolName, targetApp := range uniqueTargetApps {
		networkPolicy := networkingv1.NetworkPolicy{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: routing.Namespace,
				Name:      netpolName,
			},
		}

		_, err = ctrlutil.CreateOrPatch(ctx, r.GetClient(), &networkPolicy, func() error {
			applicationNamespacedName := types.NamespacedName{Namespace: routing.Namespace, Name: targetApp}
			targetApplication, err := getApplication(r.GetClient(), ctx, applicationNamespacedName)
			if err != nil {
				return err
			}

			err = ctrlutil.SetControllerReference(routing, &networkPolicy, r.GetScheme())
			if err != nil {
				return err
			}

			networkPolicy.Spec = networkingv1.NetworkPolicySpec{
				PodSelector: metav1.LabelSelector{
					MatchLabels: util.GetPodAppSelector(targetApp),
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
									MatchLabels: util.GetIstioGatewayLabelSelector(routing.GetIsInternal(), routing.Spec.Hostname),
								},
							},
						},
						Ports: []networkingv1.NetworkPolicyPort{
							{
								Port: util.PointTo(intstr.FromInt32(int32(targetApplication.Spec.Port))),
							},
						},
					},
				},
			}
			util.SetCommonAnnotations(&networkPolicy)
			return nil
		})
	}
	if err != nil {
		err = r.setConditionNetworkPolicySynced(ctx, routing, ConditionStatusFalse, err.Error())
		return util.RequeueWithError(err)
	}

	// Delete network policies that are not defined by routing resource anymore
	networPolicyInNamespace := networkingv1.NetworkPolicyList{}
	err = r.GetClient().List(ctx, &networPolicyInNamespace, client.InNamespace(routing.Namespace))
	if err != nil {
		return util.RequeueWithError(err)
	}

	var networkPoliciesToDelete []networkingv1.NetworkPolicy
	for _, networkPolicy := range networPolicyInNamespace.Items {
		ownerIndex := slices.IndexFunc(networkPolicy.GetOwnerReferences(), func(ownerReference metav1.OwnerReference) bool {
			return ownerReference.Name == routing.Name
		})
		networkPolicyOwnedByThisApplication := ownerIndex != -1
		if !networkPolicyOwnedByThisApplication {
			continue
		}

		_, ok := uniqueTargetApps[networkPolicy.Name]
		if ok {
			continue
		}

		networkPoliciesToDelete = append(networkPoliciesToDelete, networkPolicy)
	}

	for _, networkPolicy := range networkPoliciesToDelete {
		err = r.GetClient().Delete(ctx, &networkPolicy)
		err = client.IgnoreNotFound(err)
		if err != nil {
			err = r.setConditionNetworkPolicySynced(ctx, routing, ConditionStatusFalse, err.Error())
			return util.RequeueWithError(err)
		}
	}

	err = r.setConditionNetworkPolicySynced(ctx, routing, ConditionStatusTrue, ConditionMessageNetworkPolicySynced)
	return util.RequeueWithError(err)
}

func getNetworkPolicyName(routing *skiperatorv1alpha1.Routing, targetApp string) string {
	return fmt.Sprintf("%s-%s-istio-ingress", routing.Name, targetApp)
}

func getApplication(client client.Client, ctx context.Context, namespacedName types.NamespacedName) (skiperatorv1alpha1.Application, error) {
	application := skiperatorv1alpha1.Application{}

	err := client.Get(ctx, namespacedName, &application)

	return application, err
}
