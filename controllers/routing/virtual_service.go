package routingcontroller

import (
	"context"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	networkingv1beta1api "istio.io/api/networking/v1beta1"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Generate(ctx context.Context, routing *skiperatorv1alpha1.Routing) *networkingv1beta1.VirtualService {
	virtualService := networkingv1beta1.VirtualService{
		ObjectMeta: v1.ObjectMeta{
			Name:      routing.GetVirtualServiceName(),
			Namespace: routing.Namespace,
		},
	}

	var err error

	virtualService.Spec = networkingv1beta1api.VirtualService{
		ExportTo: []string{".", "istio-system", "istio-gateways"},
		Gateways: []string{
			routing.GetGatewayName(),
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
						Host: route.TargetApp,
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
}
