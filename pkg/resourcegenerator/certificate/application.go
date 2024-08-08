package certificate

import (
	"fmt"
	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	v1 "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/reconciliation"
	"github.com/kartverket/skiperator/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

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
	for _, h := range hosts {
		if h.UsesCustomCert() {
			continue
		}
		certificateName := fmt.Sprintf("%s-%s-ingress-%x", application.Namespace, application.Name, util.GenerateHashFromName(h.Hostname))
		certificate := certmanagerv1.Certificate{ObjectMeta: metav1.ObjectMeta{Namespace: "istio-gateways", Name: certificateName}}

		certificate.Spec = certmanagerv1.CertificateSpec{
			IssuerRef: v1.ObjectReference{
				Kind: "ClusterIssuer",
				Name: "cluster-issuer", // Name defined in https://github.com/kartverket/certificate-management/blob/main/clusterissuer.tf
			},
			DNSNames:   []string{h.Hostname},
			SecretName: certificateName,
		}
		var obj client.Object = &certificate
		r.AddResource(obj)
	}
	ctxLog.Debug("Finished generating certificates for application", "application", application.Name)
	return nil
}
