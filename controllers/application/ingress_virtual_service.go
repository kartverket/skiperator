package applicationcontroller

import (
	"context"
	"fmt"
	"hash/fnv"

	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/util"
	networkingv1beta1api "istio.io/api/networking/v1beta1"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var ExportToNamespaces = []string{".", "istio-system", "istio-gateways"}

func (r *ApplicationReconciler) reconcileIngressVirtualService(ctx context.Context, application *skiperatorv1alpha1.Application) (reconcile.Result, error) {
	controllerName := "IngressVirtualService"
	r.SetControllerProgressing(ctx, application, controllerName)

	commonVirtualService, err := r.defineCommonVirtualService(ctx, application)
	if err != nil {
		r.SetControllerError(ctx, application, controllerName, err)
		return reconcile.Result{}, err
	}

	result, err := r.createOrUpdateVirtualService(ctx, *application, *commonVirtualService)
	if err != nil {
		return result, err
	}

	redirectVirtualService, err := r.defineRedirectVirtualService(ctx, application)
	if err != nil {
		r.SetControllerError(ctx, application, controllerName, err)
		return reconcile.Result{}, err
	}

	if application.Spec.RedirectToHTTPS {
		result, err := r.createOrUpdateVirtualService(ctx, *application, *redirectVirtualService)
		if err != nil {
			return result, err
		}
	} else {
		err = r.GetClient().Delete(ctx, redirectVirtualService)
		err = client.IgnoreNotFound(err)
		if err != nil {
			r.SetControllerError(ctx, application, controllerName, err)
			return reconcile.Result{}, err
		}
	}

	if len(application.Spec.Ingresses) == 0 {
		err = r.GetClient().Delete(ctx, commonVirtualService)
		err = client.IgnoreNotFound(err)
		if err != nil {
			r.SetControllerError(ctx, application, controllerName, err)
			return reconcile.Result{}, err
		}

		err = r.GetClient().Delete(ctx, redirectVirtualService)
		err = client.IgnoreNotFound(err)
		if err != nil {
			r.SetControllerError(ctx, application, controllerName, err)
			return reconcile.Result{}, err
		}
	}

	r.SetControllerFinishedOutcome(ctx, application, controllerName, err)

	return reconcile.Result{}, err
}

func (r *ApplicationReconciler) defineRedirectVirtualService(ctx context.Context, application *skiperatorv1alpha1.Application) (*networkingv1beta1.VirtualService, error) {
	virtualService := networkingv1beta1.VirtualService{
		ObjectMeta: v1.ObjectMeta{
			Name:      application.Name + "-http-redirect",
			Namespace: application.Namespace,
		},
		Spec: networkingv1beta1api.VirtualService{
			ExportTo: ExportToNamespaces,
			Gateways: r.getGatewaysFromApplication(application),
			Hosts:    []string{"*"},
			Http: []*networkingv1beta1api.HTTPRoute{
				{
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
				},
			},
		},
	}

	err := ctrlutil.SetControllerReference(application, &virtualService, r.GetScheme())
	if err != nil {
		r.SetControllerError(ctx, application, controllerName, err)
		return &virtualService, err
	}

	r.SetLabelsFromApplication(ctx, &virtualService, *application)
	util.SetCommonAnnotations(&virtualService)

	return &virtualService, err
}

func (r *ApplicationReconciler) defineCommonVirtualService(ctx context.Context, application *skiperatorv1alpha1.Application) (*networkingv1beta1.VirtualService, error) {
	virtualService := networkingv1beta1.VirtualService{
		ObjectMeta: v1.ObjectMeta{
			Name:      application.Name + "-ingress",
			Namespace: application.Namespace,
		},
		Spec: networkingv1beta1api.VirtualService{
			ExportTo: ExportToNamespaces,
			Gateways: r.getGatewaysFromApplication(application),
			Hosts:    application.Spec.Ingresses,
			Http: []*networkingv1beta1api.HTTPRoute{
				{
					Route: []*networkingv1beta1api.HTTPRouteDestination{
						{
							Destination: &networkingv1beta1api.Destination{
								Host: application.Name,
							},
						},
					},
				},
			},
		},
	}

	err := ctrlutil.SetControllerReference(application, &virtualService, r.GetScheme())
	if err != nil {
		r.SetControllerError(ctx, application, controllerName, err)
		return &virtualService, err
	}

	r.SetLabelsFromApplication(ctx, &virtualService, *application)
	util.SetCommonAnnotations(&virtualService)

	return &virtualService, err
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

func (r *ApplicationReconciler) createOrUpdateVirtualService(ctx context.Context, application skiperatorv1alpha1.Application, wantedVirtualService networkingv1beta1.VirtualService) (reconcile.Result, error) {
	currentVirtualService := networkingv1beta1.VirtualService{}
	err := r.GetClient().Get(ctx, types.NamespacedName{Namespace: wantedVirtualService.Namespace, Name: wantedVirtualService.Name}, &currentVirtualService)

	if errors.IsNotFound(err) {
		err = r.GetClient().Create(ctx, &wantedVirtualService)
		if err != nil {
			r.SetControllerError(ctx, &application, controllerName, err)
			return reconcile.Result{}, err
		}
	} else if err != nil {
		r.SetControllerError(ctx, &application, controllerName, err)
		return reconcile.Result{}, err
	} else {
		err = r.GetClient().Patch(ctx, &wantedVirtualService, client.Merge)
		if err != nil {
			r.SetControllerError(ctx, &application, controllerName, err)
			return reconcile.Result{}, err
		}
	}

	return reconcile.Result{}, err
}
