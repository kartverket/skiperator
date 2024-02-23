package routingcontroller

import (
	"context"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/util"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *RoutingReconciler) reconcileNetworkPolicy(ctx context.Context, routing *skiperatorv1alpha1.Routing) (reconcile.Result, error) {
	// Set up networkpolicy to allow traffic from istio gateways to each of the services defined by the applications
	//egressServices, err := r.GetEgressServices(ctx, skipJob, skipJob.Spec.Container.AccessPolicy)
	//if err != nil {
	//	return util.RequeueWithError(err)
	//}

	//namespaces, err := r.GetNamespaces(ctx, routing)
	//if err != nil {
	//	return util.RequeueWithError(err)
	//}
	var err error

	for _, route := range routing.Spec.Routes {
		netpolName := route.TargetApp + "-istio-route"

		networkPolicy := networkingv1.NetworkPolicy{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: routing.Namespace,
				Name:      netpolName,
			},
		}

		_, err = ctrlutil.CreateOrPatch(ctx, r.GetClient(), &networkPolicy, func() error {
			var err error
			applicationNamespacedName := types.NamespacedName{Namespace: routing.Namespace, Name: route.TargetApp}
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

	return util.RequeueWithError(err)
}

func getApplication(client client.Client, ctx context.Context, namespacedName types.NamespacedName) (skiperatorv1alpha1.Application, error) {
	application := skiperatorv1alpha1.Application{}

	err := client.Get(ctx, namespacedName, &application)

	return application, err
}
