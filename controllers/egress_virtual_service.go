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

//+kubebuilder:rbac:groups=skiperator.kartverket.no,resources=applications,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=networking.istio.io,resources=virtualservices,verbs=get;list;watch;create;update;patch;delete

type EgressVirtualServiceReconciler struct {
	client client.Client
	scheme *runtime.Scheme
}

func (r *EgressVirtualServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.client = mgr.GetClient()
	r.scheme = mgr.GetScheme()

	return ctrl.NewControllerManagedBy(mgr).
		For(&skiperatorv1alpha1.Application{}).
		Owns(&networkingv1beta1.VirtualService{}, builder.WithPredicates(
			matchesPredicate[*networkingv1beta1.VirtualService](isEgressVirtualService),
		)).
		Complete(r)
}

func (r *EgressVirtualServiceReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	// Fetch application and fill defaults
	application := skiperatorv1alpha1.Application{}
	err := r.client.Get(ctx, req.NamespacedName, &application)
	if err != nil {
		err = client.IgnoreNotFound(err)
		return reconcile.Result{}, err
	}
	application.FillDefaults()

	// Keep track of active virtual services
	active := make(map[string]struct{}, len(application.Spec.AccessPolicy.Outbound.External))

	for _, rule := range application.Spec.AccessPolicy.Outbound.External {
		// Generate virtual service name
		hash := fnv.New64()
		_, _ = hash.Write([]byte(rule.Host))
		name := fmt.Sprintf("%s-egress-%x", req.Name, hash.Sum64())
		active[name] = struct{}{}

		virtualService := networkingv1beta1.VirtualService{ObjectMeta: metav1.ObjectMeta{Namespace: req.Namespace, Name: name}}
		_, err = ctrlutil.CreateOrPatch(ctx, r.client, &virtualService, func() error {
			// Set application as owner of the virtual service
			err = ctrlutil.SetControllerReference(&application, &virtualService, r.scheme)
			if err != nil {
				return err
			}

			// Avoid leaking virtual service to other namespaces
			virtualService.Spec.ExportTo = []string{".", "istio-system"}
			virtualService.Spec.Gateways = []string{"mesh", name}
			virtualService.Spec.Hosts = []string{rule.Host}

			httpCount := 0
			httpsCount := 0
			tcpCount := 0
			for _, port := range rule.Ports {
				if port.Protocol == "HTTP" {
					httpCount++
				} else if port.Protocol == "HTTPS" {
					httpsCount++
				} else if port.Protocol == "TCP" {
					tcpCount++
				} else {
					return fmt.Errorf("invalid protocol: %s", port.Protocol)
				}
			}

			virtualService.Spec.Http = make([]*networkingv1beta1api.HTTPRoute, 0, httpCount*2)
			virtualService.Spec.Tls = make([]*networkingv1beta1api.TLSRoute, 0, httpsCount*2)
			virtualService.Spec.Tcp = make([]*networkingv1beta1api.TCPRoute, 0, tcpCount*2)

			for _, port := range rule.Ports {
				if port.Protocol == "HTTP" {
					http := &networkingv1beta1api.HTTPRoute{}
					virtualService.Spec.Http = append(virtualService.Spec.Http, http)

					http.Match = make([]*networkingv1beta1api.HTTPMatchRequest, 1)
					http.Match[0] = &networkingv1beta1api.HTTPMatchRequest{}
					http.Match[0].Gateways = []string{"mesh"}
					http.Route = make([]*networkingv1beta1api.HTTPRouteDestination, 1)
					http.Route[0] = &networkingv1beta1api.HTTPRouteDestination{}
					http.Route[0].Destination = &networkingv1beta1api.Destination{}
					http.Route[0].Destination.Host = "egress-external.istio-system.svc.cluster.local"
					http.Route[0].Destination.Port = &networkingv1beta1api.PortSelector{}
					http.Route[0].Destination.Port.Number = uint32(port.Port)

					http = &networkingv1beta1api.HTTPRoute{}
					virtualService.Spec.Http = append(virtualService.Spec.Http, http)

					http.Match = make([]*networkingv1beta1api.HTTPMatchRequest, 1)
					http.Match[0] = &networkingv1beta1api.HTTPMatchRequest{}
					http.Match[0].Gateways = []string{name}
					http.Route = make([]*networkingv1beta1api.HTTPRouteDestination, 1)
					http.Route[0] = &networkingv1beta1api.HTTPRouteDestination{}
					http.Route[0].Destination = &networkingv1beta1api.Destination{}
					http.Route[0].Destination.Host = rule.Host
					http.Route[0].Destination.Port = &networkingv1beta1api.PortSelector{}
					http.Route[0].Destination.Port.Number = uint32(port.Port)
				} else if port.Protocol == "HTTPS" {
					tls := &networkingv1beta1api.TLSRoute{}
					virtualService.Spec.Tls = append(virtualService.Spec.Tls, tls)

					tls.Match = make([]*networkingv1beta1api.TLSMatchAttributes, 1)
					tls.Match[0] = &networkingv1beta1api.TLSMatchAttributes{}
					tls.Match[0].Gateways = []string{"mesh"}
					tls.Route = make([]*networkingv1beta1api.RouteDestination, 1)
					tls.Route[0] = &networkingv1beta1api.RouteDestination{}
					tls.Route[0].Destination = &networkingv1beta1api.Destination{}
					tls.Route[0].Destination.Host = "egress-external.istio-system.svc.cluster.local"
					tls.Route[0].Destination.Port = &networkingv1beta1api.PortSelector{}
					tls.Route[0].Destination.Port.Number = uint32(port.Port)

					tls = &networkingv1beta1api.TLSRoute{}
					virtualService.Spec.Tls = append(virtualService.Spec.Tls, tls)

					tls.Match = make([]*networkingv1beta1api.TLSMatchAttributes, 1)
					tls.Match[0] = &networkingv1beta1api.TLSMatchAttributes{}
					tls.Match[0].Gateways = []string{name}
					tls.Route = make([]*networkingv1beta1api.RouteDestination, 1)
					tls.Route[0] = &networkingv1beta1api.RouteDestination{}
					tls.Route[0].Destination = &networkingv1beta1api.Destination{}
					tls.Route[0].Destination.Host = rule.Host
					tls.Route[0].Destination.Port = &networkingv1beta1api.PortSelector{}
					tls.Route[0].Destination.Port.Number = uint32(port.Port)
				} else if port.Protocol == "TCP" {
					tcp := &networkingv1beta1api.TCPRoute{}
					virtualService.Spec.Tcp = append(virtualService.Spec.Tcp, tcp)

					tcp.Match = make([]*networkingv1beta1api.L4MatchAttributes, 1)
					tcp.Match[0] = &networkingv1beta1api.L4MatchAttributes{}
					tcp.Match[0].Gateways = []string{"mesh"}
					tcp.Route = make([]*networkingv1beta1api.RouteDestination, 1)
					tcp.Route[0] = &networkingv1beta1api.RouteDestination{}
					tcp.Route[0].Destination = &networkingv1beta1api.Destination{}
					tcp.Route[0].Destination.Host = "egress-external.istio-system.svc.cluster.local"
					tcp.Route[0].Destination.Port = &networkingv1beta1api.PortSelector{}
					tcp.Route[0].Destination.Port.Number = uint32(port.Port)

					tcp = &networkingv1beta1api.TCPRoute{}
					virtualService.Spec.Tcp = append(virtualService.Spec.Tcp, tcp)

					tcp.Match = make([]*networkingv1beta1api.L4MatchAttributes, 1)
					tcp.Match[0] = &networkingv1beta1api.L4MatchAttributes{}
					tcp.Match[0].Gateways = []string{name}
					tcp.Route = make([]*networkingv1beta1api.RouteDestination, 1)
					tcp.Route[0] = &networkingv1beta1api.RouteDestination{}
					tcp.Route[0].Destination = &networkingv1beta1api.Destination{}
					tcp.Route[0].Destination.Host = rule.Host
					tcp.Route[0].Destination.Port = &networkingv1beta1api.PortSelector{}
					tcp.Route[0].Destination.Port.Number = uint32(port.Port)
				} else {
					panic("should never reach here")
				}
			}

			return nil
		})
		if err != nil {
			return reconcile.Result{}, err
		}
	}

	// Clear out unused virtual services
	virtualServices := networkingv1beta1.VirtualServiceList{}
	err = r.client.List(ctx, &virtualServices, client.InNamespace(req.Namespace))
	if err != nil {
		return reconcile.Result{}, err
	}

	for i := range virtualServices.Items {
		virtualService := &virtualServices.Items[i]

		// Skip unrelated virtual services
		if !isEgressVirtualService(virtualService) {
			continue
		}

		// Skip active virtual services
		_, ok := active[virtualService.Name]
		if ok {
			continue
		}

		// Delete the rest
		err = r.client.Delete(ctx, virtualService)
		err = client.IgnoreNotFound(err)
		if err != nil {
			return reconcile.Result{}, err
		}
	}

	return reconcile.Result{}, nil
}

// Filter for virtual services named like *-egress-*
func isEgressVirtualService(virtualService *networkingv1beta1.VirtualService) bool {
	segments := strings.Split(virtualService.Name, "-")

	if len(segments) != 3 {
		return false
	}

	return segments[1] == "egress"
}
