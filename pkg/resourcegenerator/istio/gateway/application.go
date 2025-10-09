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
	multiGenerator.Register(reconciliation.ApplicationType, generateForApplication)
}

func generateForApplication(r reconciliation.Reconciliation) error {
	ctxLog := r.GetLogger()
	ctxLog.Debug("Attempting to generate gateway for application", "objectname", r.GetSKIPObject().GetName())

	if r.GetType() != reconciliation.ApplicationType {
		return fmt.Errorf("gateway only supports Application type")
	}

	application, ok := r.GetSKIPObject().(*skiperatorv1beta1.Application)
	if !ok {
		return fmt.Errorf("failed to cast object to Application")
	}

	hosts, err := application.Spec.Hosts()
	if err != nil {
		return fmt.Errorf("failure to get hosts from application: %w", err)
	}

	// Generate separate gateway for each ingress
	for _, h := range hosts.AllHosts() {
		name := fmt.Sprintf("%s-ingress-%x", application.Name, util.GenerateHashFromName(h.Hostname))
		gateway := networkingv1.Gateway{ObjectMeta: metav1.ObjectMeta{Namespace: application.Namespace, Name: name}}

		gateway.Spec.Selector = util.GetIstioGatewayLabelSelector(h.Hostname)

		gatewayServersToAdd := []*networkingv1api.Server{}

		baseHttpGatewayServer := &networkingv1api.Server{
			Hosts: []string{h.Hostname},
			Port: &networkingv1api.Port{
				Number:   80,
				Name:     "http",
				Protocol: "HTTP",
			},
		}

		determinedCredentialName := application.Namespace + "-" + name
		if h.UsesCustomCert() {
			determinedCredentialName = *h.CustomCertificateSecret
		}

		httpsGatewayServer := &networkingv1api.Server{
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
		}

		gatewayServersToAdd = append(gatewayServersToAdd, baseHttpGatewayServer, httpsGatewayServer)

		gateway.Spec.Servers = gatewayServersToAdd
		r.AddResource(&gateway)
	}

	ctxLog.Debug("Finished generating ingress gateways for application", "application", application.Name)
	return nil
}
