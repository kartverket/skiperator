package applicationcontroller

import (
	"context"
	"fmt"
	"hash/fnv"
	"regexp"

	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/util"
	networkingv1beta1api "istio.io/api/networking/v1beta1"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *ApplicationReconciler) reconcileIngressVirtualService(ctx context.Context, application *skiperatorv1alpha1.Application) (reconcile.Result, error) {
	controllerName := "IngressVirtualService"
	r.SetControllerProgressing(ctx, application, controllerName)

	var err error
	virtualService := networkingv1beta1.VirtualService{ObjectMeta: metav1.ObjectMeta{Namespace: application.Namespace, Name: application.Name + "-ingress"}}
	if len(application.Spec.Ingresses) > 0 {
		_, err = ctrlutil.CreateOrPatch(ctx, r.GetClient(), &virtualService, func() error {
			// Set application as owner of the virtual service
			err = ctrlutil.SetControllerReference(application, &virtualService, r.GetScheme())
			if err != nil {
				r.SetControllerError(ctx, application, controllerName, err)
				return err
			}

			r.SetLabelsFromApplication(ctx, &virtualService, *application)
			util.SetCommonAnnotations(&virtualService)

			gateways := make([]string, 0, len(application.Spec.Ingresses))
			for _, hostname := range application.Spec.Ingresses {
				// Generate gateway name
				hash := fnv.New64()
				_, _ = hash.Write([]byte(hostname))
				name := fmt.Sprintf("%s-ingress-%x", application.Name, hash.Sum64())
				gateways = append(gateways, name)
			}

			// Avoid leaking virtual service to other namespaces
			virtualService.Spec.ExportTo = []string{".", "istio-system", "istio-gateways"}
			virtualService.Spec.Gateways = gateways
			virtualService.Spec.Hosts = application.Spec.Ingresses

			virtualService.Spec.Http = make([]*networkingv1beta1api.HTTPRoute, 1)
			virtualService.Spec.Http[0] = &networkingv1beta1api.HTTPRoute{}
			virtualService.Spec.Http[0].Route = make([]*networkingv1beta1api.HTTPRouteDestination, 1)
			virtualService.Spec.Http[0].Route[0] = &networkingv1beta1api.HTTPRouteDestination{}
			virtualService.Spec.Http[0].Route[0].Destination = &networkingv1beta1api.Destination{}
			virtualService.Spec.Http[0].Route[0].Destination.Host = application.Name

			if application.Spec.RedirectIngresses {
				virtualService.Spec.Http[0].Redirect.Scheme = "https"
			}

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

	r.SetControllerFinishedOutcome(ctx, application, controllerName, err)

	return reconcile.Result{}, err
}

// Filter for virtual services named like *-ingress
func isIngressVirtualService(virtualService *networkingv1beta1.VirtualService) bool {
	match, _ := regexp.MatchString("^.*-ingress$", virtualService.Name)
	return match
}
