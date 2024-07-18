package certificate

import (
	"fmt"
	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	v1 "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/reconciliation"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/resourceutils"
	"github.com/kartverket/skiperator/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func generateForApplication(r reconciliation.Reconciliation) error {
	ctxLog := r.GetLogger()
	ctxLog.Debug("Attempting to generate certificates for application", "application", r.GetReconciliationObject().GetName())

	if r.GetType() != reconciliation.ApplicationType {
		return fmt.Errorf("certificate only supports Application type")
	}

	application, ok := r.GetReconciliationObject().(*skiperatorv1alpha1.Application)
	if !ok {
		return fmt.Errorf("failed to cast object to Application")
	}

	// Generate separate cert for each ingress
	for _, hostname := range application.Spec.Ingresses {
		certificateName := fmt.Sprintf("%s-%s-ingress-%x", application.Namespace, application.Name, util.GenerateHashFromName(hostname))

		certificate := certmanagerv1.Certificate{ObjectMeta: metav1.ObjectMeta{Namespace: "istio-gateways", Name: certificateName}}

		resourceutils.SetApplicationLabels(&certificate, application)

		certificate.Spec = certmanagerv1.CertificateSpec{
			IssuerRef: v1.ObjectReference{
				Kind: "ClusterIssuer",
				Name: "cluster-issuer", // Name defined in https://github.com/kartverket/certificate-management/blob/main/clusterissuer.tf
			},
			DNSNames:   []string{hostname},
			SecretName: certificateName,
		}
		var obj client.Object = &certificate
		r.AddResource(&obj)
	}
	ctxLog.Debug("Finished generating certificates for application", "application", application.Name)
	return nil
}
