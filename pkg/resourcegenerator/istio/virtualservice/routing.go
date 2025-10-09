package virtualservice

import (
	"fmt"

	skiperatorv1beta1 "github.com/kartverket/skiperator/api/v1beta1"
	"github.com/kartverket/skiperator/pkg/reconciliation"
	networkingv1api "istio.io/api/networking/v1"
	networkingv1 "istio.io/client-go/pkg/apis/networking/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func init() {
	multiGenerator.Register(reconciliation.RoutingType, generateForRouting)
}

func generateForRouting(r reconciliation.Reconciliation) error {
	ctxLog := r.GetLogger()
	ctxLog.Debug("Attempting to generate virtual service for routing", "routing", r.GetSKIPObject().GetName())

	routing, ok := r.GetSKIPObject().(*skiperatorv1beta1.Routing)
	if !ok {
		return fmt.Errorf("failed to cast object to Application")
	}

	virtualService := networkingv1.VirtualService{
		ObjectMeta: v1.ObjectMeta{
			Name:      routing.GetVirtualServiceName(),
			Namespace: routing.Namespace,
		},
	}

	h, err := routing.Spec.GetHost()
	if err != nil || h == nil {
		return fmt.Errorf("failed to get host from routing: %w", err)
	}

	virtualService.Spec = networkingv1api.VirtualService{
		ExportTo: []string{".", "istio-system", "istio-gateways"},
		Gateways: []string{
			routing.GetGatewayName(),
		},
		Hosts: []string{
			h.Hostname,
		},
		Http: []*networkingv1api.HTTPRoute{},
	}

	if routing.GetRedirectToHTTPS() {
		virtualService.Spec.Http = append(virtualService.Spec.Http, &networkingv1api.HTTPRoute{
			Name: "redirect-to-https",
			Match: []*networkingv1api.HTTPMatchRequest{
				{
					Port: 80,
				},
			},
			Redirect: &networkingv1api.HTTPRedirect{
				Scheme:       "https",
				RedirectCode: 308,
			},
		})
	}

	for _, route := range routing.Spec.Routes {

		httpRoute := &networkingv1api.HTTPRoute{
			Name: route.TargetApp,
			Match: []*networkingv1api.HTTPMatchRequest{
				{
					Port: 443,
					Uri: &networkingv1api.StringMatch{
						MatchType: &networkingv1api.StringMatch_Prefix{
							Prefix: route.PathPrefix,
						},
					},
				},
			},
			Route: []*networkingv1api.HTTPRouteDestination{
				{
					Destination: &networkingv1api.Destination{
						Host: route.TargetApp,
						Port: &networkingv1api.PortSelector{
							Number: uint32(route.Port),
						},
					},
				},
			},
		}

		if route.RewriteUri {
			httpRoute.Rewrite = &networkingv1api.HTTPRewrite{
				Uri: "/",
			}
		}

		virtualService.Spec.Http = append(virtualService.Spec.Http, httpRoute)
	}
	r.AddResource(&virtualService)
	ctxLog.Debug("Finished generating virtual service for routing", "routing", routing.Name)
	return nil
}
