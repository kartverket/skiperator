package virtualservice

import (
	"fmt"
	"hash/fnv"

	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/reconciliation"
	networkingv1beta1api "istio.io/api/networking/v1beta1"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func init() {
	multiGenerator.Register(reconciliation.ApplicationType, generateForApplication)
}

func generateForApplication(r reconciliation.Reconciliation) error {
	ctxLog := r.GetLogger()
	ctxLog.Debug("Attempting to generate virtual service for application", "application", r.GetSKIPObject().GetName())

	application, ok := r.GetSKIPObject().(*skiperatorv1alpha1.Application)
	if !ok {
		return fmt.Errorf("failed to cast object to Application")
	}

	virtualService := networkingv1beta1.VirtualService{
		ObjectMeta: v1.ObjectMeta{
			Name:      application.Name + "-ingress",
			Namespace: application.Namespace,
		},
	}

	hosts, err := application.Spec.Hosts()
	if err != nil {
		return err
	}

	if len(hosts.Hostnames()) > 0 {
		virtualService.Spec = networkingv1beta1api.VirtualService{
			ExportTo: []string{".", "istio-system", "istio-gateways"},
			Gateways: getGatewaysFromApplication(application),
			Hosts:    hosts.Hostnames(),
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
						Port: &networkingv1beta1api.PortSelector{
							Number: uint32(application.Spec.Port),
						},
					},
				},
			},
		})
		r.AddResource(&virtualService)
		ctxLog.Debug("Added virtual service to application", "application", application.Name)
	}

	ctxLog.Debug("Finished generating virtual service for application", "application", application.Name)
	return nil
}

func getGatewaysFromApplication(application *skiperatorv1alpha1.Application) []string {
	hosts, _ := application.Spec.Hosts()
	gateways := make([]string, 0, hosts.Count())
	for _, hostname := range hosts.Hostnames() {
		// Generate gateway name
		hash := fnv.New64()
		_, _ = hash.Write([]byte(hostname))
		name := fmt.Sprintf("%s-ingress-%x", application.Name, hash.Sum64())
		gateways = append(gateways, name)
	}

	return gateways
}
