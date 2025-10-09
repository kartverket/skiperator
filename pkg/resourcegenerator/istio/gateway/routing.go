package gateway

import (
	"fmt"

	skiperatorv1beta1 "github.com/kartverket/skiperator/api/v1beta1"
	"github.com/kartverket/skiperator/pkg/reconciliation"
	"github.com/kartverket/skiperator/pkg/util"
	networkingv1api "istio.io/api/networking/v1"
	networkingv1 "istio.io/client-go/pkg/apis/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func init() {
	multiGenerator.Register(reconciliation.RoutingType, generateForRouting)
}

func generateForRouting(r reconciliation.Reconciliation) error {
	ctxLog := r.GetLogger()
	ctxLog.Debug("Attempting to generate gateway for routing", "routing", r.GetSKIPObject().GetName())

	if r.GetType() != reconciliation.RoutingType {
		return fmt.Errorf("gateway only supports routing type")
	}
	routing, ok := r.GetSKIPObject().(*skiperatorv1beta1.Routing)
	if !ok {
		return fmt.Errorf("failed to cast object to routing")
	}

	h, err := routing.Spec.GetHost()
	if err != nil {
		return err
	}

	gateway := networkingv1.Gateway{ObjectMeta: metav1.ObjectMeta{Namespace: routing.Namespace, Name: routing.GetGatewayName()}}

	var determinedCredentialName string
	if h.UsesCustomCert() {
		determinedCredentialName = *h.CustomCertificateSecret
	} else {
		determinedCredentialName, err = routing.GetCertificateName()
		if err != nil {
			return err
		}
	}

	gateway.Spec.Selector = util.GetIstioGatewayLabelSelector(h.Hostname)
	gateway.Spec.Servers = []*networkingv1api.Server{
		{
			Hosts: []string{h.Hostname},
			Port: &networkingv1api.Port{
				Number:   80,
				Name:     "http",
				Protocol: "HTTP",
			},
		},
		{
			Hosts: []string{h.Hostname},
			Port: &networkingv1api.Port{
				Number:   443,
				Name:     "https",
				Protocol: "HTTPS",
			},
			Tls: &networkingv1api.ServerTLSSettings{
				Mode:           networkingv1api.ServerTLSSettings_SIMPLE,
				CredentialName: determinedCredentialName,
			},
		},
	}

	r.AddResource(&gateway)

	ctxLog.Debug("Finished generating ingress gateways for routing", "routing", routing.Name)
	return nil

}
