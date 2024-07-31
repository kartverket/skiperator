package gateway

import (
	"fmt"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/reconciliation"
	"github.com/kartverket/skiperator/pkg/util"
	networkingv1beta1api "istio.io/api/networking/v1beta1"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func generateForRouting(r reconciliation.Reconciliation) error {
	ctxLog := r.GetLogger()
	ctxLog.Debug("Attempting to generate gateway for routing", "routing", r.GetSKIPObject().GetName())

	if r.GetType() != reconciliation.RoutingType {
		return fmt.Errorf("gateway only supports routing type")
	}
	routing, ok := r.GetSKIPObject().(*skiperatorv1alpha1.Routing)
	if !ok {
		return fmt.Errorf("failed to cast object to routing")
	}

	h, err := routing.Spec.GetHost()
	if err != nil {
		return err
	}

	gateway := networkingv1beta1.Gateway{ObjectMeta: metav1.ObjectMeta{Namespace: routing.Namespace, Name: routing.GetGatewayName()}}

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
	gateway.Spec.Servers = []*networkingv1beta1api.Server{
		{
			Hosts: []string{h.Hostname},
			Port: &networkingv1beta1api.Port{
				Number:   80,
				Name:     "http",
				Protocol: "HTTP",
			},
		},
		{
			Hosts: []string{h.Hostname},
			Port: &networkingv1beta1api.Port{
				Number:   443,
				Name:     "https",
				Protocol: "HTTPS",
			},
			Tls: &networkingv1beta1api.ServerTLSSettings{
				Mode:           networkingv1beta1api.ServerTLSSettings_SIMPLE,
				CredentialName: determinedCredentialName,
			},
		},
	}

	var obj client.Object = &gateway
	r.AddResource(obj)

	ctxLog.Debug("Finished generating ingress gateways for routing", "routing", routing.Name)
	return nil

}
