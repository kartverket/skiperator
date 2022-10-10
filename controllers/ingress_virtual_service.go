package controllers

import (
	"context"
	"fmt"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"hash/fnv"
	networkingv1beta1api "istio.io/api/networking/v1beta1"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"strings"
)

//+kubebuilder:rbac:groups=skiperator.kartverket.no,resources=applications,verbs=get;list;watch
//+kubebuilder:rbac:groups=networking.istio.io,resources=virtualservices,verbs=get;list;watch;create;update;patch;delete

type IngressVirtualServiceReconciler struct {
	client client.Client
	scheme *runtime.Scheme
}

func (r *IngressVirtualServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.client = mgr.GetClient()
	r.scheme = mgr.GetScheme()

	return newControllerManagedBy[*skiperatorv1alpha1.Application](mgr).
		For(&skiperatorv1alpha1.Application{}).
		Owns(&networkingv1beta1.VirtualService{}, builder.WithPredicates(
			matchesPredicate[*networkingv1beta1.VirtualService](isIngressVirtualService),
		)).
		Complete(r)
}

func (r *IngressVirtualServiceReconciler) Reconcile(ctx context.Context, application *skiperatorv1alpha1.Application) (reconcile.Result, error) {
	application.FillDefaults()

	var err error
	virtualService := networkingv1beta1.VirtualService{ObjectMeta: metav1.ObjectMeta{Namespace: application.Namespace, Name: application.Name + "-ingress"}}
	if len(application.Spec.Ingresses) > 0 {
		_, err = ctrlutil.CreateOrPatch(ctx, r.client, &virtualService, func() error {
			// Set application as owner of the virtual service
			err = ctrlutil.SetControllerReference(application, &virtualService, r.scheme)
			if err != nil {
				return err
			}

			gateways := make([]string, 0, len(application.Spec.Ingresses))
			for _, hostname := range application.Spec.Ingresses {
				// Generate gateway name
				hash := fnv.New64()
				_, _ = hash.Write([]byte(hostname))
				name := fmt.Sprintf("%s-ingress-%x", application.Name, hash.Sum64())
				gateways = append(gateways, name)
			}

			// Avoid leaking virtual service to other namespaces
			virtualService.Spec.ExportTo = []string{".", "istio-system"}
			virtualService.Spec.Gateways = gateways
			virtualService.Spec.Hosts = application.Spec.Ingresses

			virtualService.Spec.Http = make([]*networkingv1beta1api.HTTPRoute, 1)
			virtualService.Spec.Http[0] = &networkingv1beta1api.HTTPRoute{}
			virtualService.Spec.Http[0].Route = make([]*networkingv1beta1api.HTTPRouteDestination, 1)
			virtualService.Spec.Http[0].Route[0] = &networkingv1beta1api.HTTPRouteDestination{}
			virtualService.Spec.Http[0].Route[0].Destination = &networkingv1beta1api.Destination{}
			virtualService.Spec.Http[0].Route[0].Destination.Host = application.Name

			return nil
		})
	} else {
		err = r.client.Delete(ctx, &virtualService)
		err = client.IgnoreNotFound(err)
	}
	return reconcile.Result{}, err
}

// Filter for virtual services named like *-ingress
func isIngressVirtualService(virtualService *networkingv1beta1.VirtualService) bool {
	segments := strings.Split(virtualService.Name, "-")

	if len(segments) != 2 {
		return false
	}

	return segments[1] == "ingress"
}
