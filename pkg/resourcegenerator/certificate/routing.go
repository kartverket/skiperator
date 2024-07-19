package certificate

import (
	"fmt"
	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	certmanagermetav1 "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/reconciliation"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func generateForRouting(r reconciliation.Reconciliation) error {
	ctxLog := r.GetLogger()
	ctxLog.Debug("Attempting to generate certificates for routing", "routing", r.GetReconciliationObject().GetName())

	if r.GetType() != reconciliation.RoutingType {
		return fmt.Errorf("certificate only supports routing type")
	}
	routing, ok := r.GetReconciliationObject().(*skiperatorv1alpha1.Routing)
	if !ok {
		return fmt.Errorf("failed to cast object to routing")
	}

	h, err := routing.Spec.GetHost()
	if err != nil {
		return err
	}

	// Do not create a new certificate when a custom certificate secret is specified
	if h.UsesCustomCert() {
		return nil
	}

	certificateName, err := routing.GetCertificateName()
	if err != nil {
		return err
	}

	certificate := certmanagerv1.Certificate{ObjectMeta: metav1.ObjectMeta{Namespace: IstioGatewayNamespace, Name: certificateName}}

	certificate.Spec = certmanagerv1.CertificateSpec{
		IssuerRef: certmanagermetav1.ObjectReference{
			Kind: "ClusterIssuer",
			Name: "cluster-issuer", // Name defined in https://github.com/kartverket/certificate-management/blob/main/clusterissuer.tf
		},
		DNSNames:   []string{h.Hostname},
		SecretName: certificateName,
	}

	var obj client.Object = &certificate
	r.AddResource(&obj)

	ctxLog.Debug("Finished generating certificates for routing", "routing", routing.Name)
	return nil
}