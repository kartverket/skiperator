package certificate

import (
	"context"
	"fmt"
	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	v1 "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/log"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/resourceutils"
	"github.com/kartverket/skiperator/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Generate(ctx context.Context, application *skiperatorv1alpha1.Application) []certmanagerv1.Certificate {
	ctxLog := log.FromContext(ctx)
	ctxLog.Debug("Attempting to create certificates for application", application.Name)

	var certificates []certmanagerv1.Certificate

	// Generate separate gateway for each ingress
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
		certificates = append(certificates, certificate)
	}
	ctxLog.Debug("Finished creating certificates for application", application.Name)
	return certificates
}
