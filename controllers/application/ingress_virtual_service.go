package applicationcontroller

import (
	"context"
	"fmt"
	"hash/fnv"

	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	networkingv1beta1api "istio.io/api/networking/v1beta1"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *ApplicationReconciler) reconcileIngressVirtualService(ctx context.Context, application *skiperatorv1alpha1.Application) (reconcile.Result, error) {
	controllerName := "IngressVirtualService"
	r.SetControllerProgressing(ctx, application, controllerName)

	virtualService := networkingv1beta1.VirtualService{
		ObjectMeta: v1.ObjectMeta{
			Name:      application.Name + "-ingress",
			Namespace: application.Namespace,
		},
	}

	var err error

	if len(application.Spec.Ingresses) > 0 {
		_, err = ctrlutil.CreateOrPatch(ctx, r.GetClient(), &virtualService, func() error {

			err := ctrlutil.SetControllerReference(application, &virtualService, r.GetScheme())
			if err != nil {
				r.SetControllerError(ctx, application, controllerName, err)
				return err
			}
			virtualService.Spec = networkingv1beta1api.VirtualService{
				ExportTo: []string{".", "istio-system", "istio-gateways"},
				Gateways: r.getGatewaysFromApplication(application),
				Hosts:    application.Spec.Ingresses,
				Http:     []*networkingv1beta1api.HTTPRoute{},
			}

			if application.Spec.RedirectToHTTPS != nil && *application.Spec.RedirectToHTTPS {
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

			virtualService.Spec.Http = append(virtualService.Spec.Http, &networkingv1beta1api.HTTPRoute{
				Name: "default-app-route",
				Route: []*networkingv1beta1api.HTTPRouteDestination{
					{
						Destination: &networkingv1beta1api.Destination{
							Host: application.Name,
						},
					},
				},
			})

			return nil
		})

	} else {
		err = r.GetClient().Delete(ctx, &virtualService)
		err = client.IgnoreNotFound(err)
		if err != nil {
			r.SetControllerError(ctx, application, controllerName, err)
			return reconcile.Result{}, err
		}
	}

	if err != nil {
		r.SetControllerError(ctx, application, controllerName, err)
		return reconcile.Result{}, err
	}

	r.SetControllerFinishedOutcome(ctx, application, controllerName, err)

	return reconcile.Result{}, err
}

func (r *ApplicationReconciler) getGatewaysFromApplication(application *skiperatorv1alpha1.Application) []string {
	gateways := make([]string, 0, len(application.Spec.Ingresses))
	for _, hostname := range application.Spec.Ingresses {
		// Generate gateway name
		hash := fnv.New64()
		_, _ = hash.Write([]byte(hostname))
		name := fmt.Sprintf("%s-ingress-%x", application.Name, hash.Sum64())
		gateways = append(gateways, name)
	}

	return gateways
}
