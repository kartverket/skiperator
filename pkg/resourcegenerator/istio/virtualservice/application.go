package virtualservice

import (
	"fmt"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/v2/api/v1alpha1"
	"github.com/kartverket/skiperator/v2/api/v1alpha1/istiotypes"
	"github.com/kartverket/skiperator/v2/pkg/reconciliation"
	"google.golang.org/protobuf/types/known/durationpb"
	"hash/fnv"
	networkingv1api "istio.io/api/networking/v1"
	networkingv1 "istio.io/client-go/pkg/apis/networking/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strconv"
	"strings"
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

	virtualService := networkingv1.VirtualService{
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
		virtualService.Spec = networkingv1api.VirtualService{
			ExportTo: []string{".", "istio-system", "istio-gateways"},
			Gateways: getGatewaysFromApplication(application),
			Hosts:    hosts.Hostnames(),
			Http:     []*networkingv1api.HTTPRoute{},
		}

		if application.Spec.RedirectToHTTPS != nil && *application.Spec.RedirectToHTTPS {
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

		virtualService.Spec.Http = append(virtualService.Spec.Http, &networkingv1api.HTTPRoute{
			Name: "default-app-route",
			Route: []*networkingv1api.HTTPRouteDestination{
				{
					Destination: &networkingv1api.Destination{
						Host: application.Name,
						Port: &networkingv1api.PortSelector{
							Number: uint32(application.Spec.Port),
						},
					},
				},
			},
			Retries: generateRetryPolicy(application.Spec.IstioSettings.Retries),
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

func generateRetryPolicy(re *istiotypes.Retries) *networkingv1api.HTTPRetry {
	conditions := "connect-failure,refused-stream,unavailable,cancelled"

	if re == nil {
		return nil
	}

	if codes := re.RetryOnHttpResponseCodes; codes != nil && len(*codes) > 0 {
		var httpRcs []string
		for _, v := range *codes {
			if v.StrVal == "" {
				httpRcs = append(httpRcs, strconv.Itoa(int(v.IntVal)))
			} else {
				httpRcs = append(httpRcs, v.StrVal)
			}
		}
		conditions = fmt.Sprintf("%s,%s", conditions, strings.Join(httpRcs, ","))
	}

	policy := &networkingv1api.HTTPRetry{
		RetryOn: conditions,
	}

	// Default to two retry attempts here, avoiding setting default in the type because we do not want issues with argo sync
	policy.Attempts = int32(2)
	if re.Attempts != nil {
		policy.Attempts = *re.Attempts
	}

	if re.PerTryTimeout != nil {
		policy.PerTryTimeout = durationpb.New(re.PerTryTimeout.Duration)
	}

	return policy
}
