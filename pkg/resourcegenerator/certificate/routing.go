package certificate

import (
	"fmt"

	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/reconciliation"
)

func init() {
	multiGenerator.Register(reconciliation.RoutingType, generateForRouting)
}

func generateForRouting(r reconciliation.Reconciliation) error {
	ctxLog := r.GetLogger()
	ctxLog.Debug("Attempting to generate certificates for routing", "routing", r.GetSKIPObject().GetName())

	if r.GetType() != reconciliation.RoutingType {
		return fmt.Errorf("certificate only supports routing type")
	}
	routing, ok := r.GetSKIPObject().(*skiperatorv1alpha1.Routing)
	if !ok {
		return fmt.Errorf("failed to cast object to routing")
	}

	h, err := routing.Spec.GetHost()
	if err != nil {
		return err
	}

	// Do not create a new certificate when a custom certificate secret is specified
	if h.UsesCustomCert() {
		ctxLog.Debug("Skipping certificate generation for routing", "routing", routing.Name, "reason", "custom certificate secret specified")
		return nil
	}

	certificateName, err := routing.GetCertificateName(h)
	if err != nil {
		return err
	}

	if r.GenerateLegacyRouting() {
		r.AddResource(newCertificate(IstioGatewayNamespace, certificateName, h.Hostname))
	}
	if routing.UsesStandardRouting() {
		r.AddResource(newCertificate(routing.Namespace, certificateName, h.Hostname))
	}

	ctxLog.Debug("Finished generating certificates for routing", "routing", routing.Name)
	return nil
}
