package routingcontroller

import (
	"context"
	"github.com/kartverket/skiperator/pkg/util"
	"k8s.io/apimachinery/pkg/types"

	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	networkingv1beta1api "istio.io/api/networking/v1beta1"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *RoutingReconciler) reconcileVirtualService(ctx context.Context, routing *skiperatorv1alpha1.Routing) (reconcile.Result, error) {
	//controllerName := "VirtualService"
	//r.SetControllerProgressing(ctx, routing, controllerName)

	virtualService := networkingv1beta1.VirtualService{
		ObjectMeta: v1.ObjectMeta{
			Name:      routing.Name + "-ingress",
			Namespace: routing.Namespace,
		},
	}

	var err error

	_, err = ctrlutil.CreateOrPatch(ctx, r.GetClient(), &virtualService, func() error {

		err := ctrlutil.SetControllerReference(routing, &virtualService, r.GetScheme())
		if err != nil {
			//r.SetControllerError(ctx, routing, controllerName, err)
			return err
		}
		virtualService.Spec = networkingv1beta1api.VirtualService{
			ExportTo: []string{".", "istio-system", "istio-gateways"},
			Gateways: []string{
				routing.Name + "-gateway",
			},
			Hosts: []string{
				routing.Spec.Hostname,
			},
			Http: []*networkingv1beta1api.HTTPRoute{},
		}

		if routing.GetRedirectToHTTPS() {
			virtualService.Spec.Http = append(virtualService.Spec.Http, &networkingv1beta1api.HTTPRoute{
				Name: "redirect-to-https",
				Match: []*networkingv1beta1api.HTTPMatchRequest{
					{
						WithoutHeaders: map[string]*networkingv1beta1api.StringMatch{
							":path": {
								MatchType: &networkingv1beta1api.StringMatch_Prefix{
									Prefix: "/.well-known/acme-challenge/",
								},
							},
						},
						Port: 80,
					},
				},
				Redirect: &networkingv1beta1api.HTTPRedirect{
					Scheme:       "https",
					RedirectCode: 308,
				},
			})
		}

		for _, route := range routing.Spec.Routes {
			applicationNamespacedName := types.NamespacedName{Namespace: routing.Namespace, Name: route.TargetApp}
			targetApplication, err := getApplication(r.GetClient(), ctx, applicationNamespacedName)
			if err != nil {
				return err
			}

			httpRoute := &networkingv1beta1api.HTTPRoute{
				Name: route.TargetApp,
				Match: []*networkingv1beta1api.HTTPMatchRequest{
					{
						Port: 443,
						Uri: &networkingv1beta1api.StringMatch{
							MatchType: &networkingv1beta1api.StringMatch_Prefix{
								Prefix: route.PathPrefix,
							},
						},
					},
				},
				Route: []*networkingv1beta1api.HTTPRouteDestination{
					{
						Destination: &networkingv1beta1api.Destination{
							Host: targetApplication.Name,
							Port: &networkingv1beta1api.PortSelector{
								Number: uint32(targetApplication.Spec.Port),
							},
						},
					},
				},
			}

			if route.RewriteUri {
				httpRoute.Rewrite = &networkingv1beta1api.HTTPRewrite{
					Uri: "/",
				}
			}

			virtualService.Spec.Http = append(virtualService.Spec.Http, httpRoute)
		}
		return nil
	})

	if err != nil {
		//r.SetControllerError(ctx, application, controllerName, err)
		return util.RequeueWithError(err)
	}

	//r.SetControllerFinishedOutcome(ctx, application, controllerName, err)

	return util.RequeueWithError(err)
}
