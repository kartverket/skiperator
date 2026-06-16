package certificate

import (
	"fmt"

	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	certmanagermetav1 "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/reconciliation"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func init() {
	multiGenerator.Register(reconciliation.ApplicationType, generateForApplication)
}

func generateForApplication(r reconciliation.Reconciliation) error {
	ctxLog := r.GetLogger()
	ctxLog.Debug("Attempting to generate certificates for application", "application", r.GetSKIPObject().GetName())

	if r.GetType() != reconciliation.ApplicationType {
		return fmt.Errorf("certificate only supports application type")
	}

	application, ok := r.GetSKIPObject().(*skiperatorv1alpha1.Application)
	if !ok {
		return fmt.Errorf("failed to cast object to Application")
	}

	hosts, err := application.Spec.Hosts()
	if err != nil {
		return fmt.Errorf("failure to get hosts from application: %w", err)
	}

	// Generate separate cert for each ingress
	for _, h := range hosts.AllHosts() {
		if h.UsesCustomCert() {
			continue
		}
		certificateName, err := application.GetCertificateName(h)
		if err != nil {
			return err
		}
		if r.GenerateLegacyRouting() {
			r.AddResource(newCertificate(IstioGatewayNamespace, certificateName, h.Hostname))
		}
		if application.UsesStandardRouting() {
			r.AddResource(newCertificate(application.Namespace, certificateName, h.Hostname))
		}
	}
	ctxLog.Debug("Finished generating certificates for application", "application", application.Name)
	return nil
}

func newCertificate(namespace string, name string, hostname string) *certmanagerv1.Certificate {
	certificate := certmanagerv1.Certificate{ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: name}}
	certificate.Spec = certmanagerv1.CertificateSpec{
		IssuerRef: certmanagermetav1.IssuerReference{
			Kind: "ClusterIssuer",
			Name: "cluster-issuer", // Name defined in https://github.com/kartverket/certificate-management/blob/main/clusterissuer.tf
		},
		DNSNames:   []string{hostname},
		SecretName: name,
	}
	return &certificate
}
