package routingcontroller

import (
	"context"
	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	certmanagermetav1 "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const IstioGatewayNamespace = "istio-gateways"

func (r *RoutingReconciler) SkiperatorRoutingCertRequests(_ context.Context, obj client.Object) []reconcile.Request {
	certificate, isCert := obj.(*certmanagerv1.Certificate)

	if !isCert {
		return nil
	}

	isSkiperatorRoutingOwned := certificate.Labels["app.kubernetes.io/managed-by"] == "skiperator" &&
		certificate.Labels["skiperator.kartverket.no/controller"] == "routing"

	requests := make([]reconcile.Request, 0)

	if isSkiperatorRoutingOwned {
		requests = append(requests, reconcile.Request{
			NamespacedName: types.NamespacedName{
				Namespace: certificate.Labels["application.skiperator.no/app-namespace"],
				Name:      certificate.Labels["application.skiperator.no/app-name"],
			},
		})
	}

	return requests
}

func (r *RoutingReconciler) reconcileCertificate(ctx context.Context, routing *skiperatorv1alpha1.Routing) (reconcile.Result, error) {
	h, err := routing.Spec.GetHost()
	if err != nil {
		err = r.setConditionCertificateSynced(ctx, routing, ConditionStatusFalse, err.Error())
		// TODO: Should we return RequeueWithError(err) here?
		return util.DoNotRequeue()
	}

	// Do not create a new certificate when a custom certificate secret is specified
	if h.CustomCertificateSecret != nil {
		err = r.setConditionCertificateSynced(ctx, routing, ConditionStatusTrue, ConditionMessageCertificateSkipped)
		return util.RequeueWithError(err)
	}

	certificateName, err := routing.GetCertificateName()
	if err != nil {
		err = r.setConditionCertificateSynced(ctx, routing, ConditionStatusFalse, err.Error())
		return util.RequeueWithError(err)
	}
	certificate := certmanagerv1.Certificate{ObjectMeta: metav1.ObjectMeta{Namespace: IstioGatewayNamespace, Name: certificateName}}

	_, err = ctrlutil.CreateOrPatch(ctx, r.GetClient(), &certificate, func() error {
		certificate.Spec = certmanagerv1.CertificateSpec{
			IssuerRef: certmanagermetav1.ObjectReference{
				Kind: "ClusterIssuer",
				Name: "cluster-issuer", // Name defined in https://github.com/kartverket/certificate-management/blob/main/clusterissuer.tf
			},
			DNSNames:   []string{h.Hostname},
			SecretName: certificateName,
		}

		certificate.Labels = getLabels(certificate, routing)

		return nil
	})
	if err != nil {
		err = r.setConditionCertificateSynced(ctx, routing, ConditionStatusFalse, err.Error())
		return util.RequeueWithError(err)
	}

	err = r.setConditionCertificateSynced(ctx, routing, ConditionStatusTrue, ConditionMessageCertificateSynced)
	return util.RequeueWithError(err)
}

func getLabels(certificate certmanagerv1.Certificate, routing *skiperatorv1alpha1.Routing) map[string]string {
	certLabels := certificate.Labels
	if len(certLabels) == 0 {
		certLabels = make(map[string]string)
	}
	certLabels["app.kubernetes.io/managed-by"] = "skiperator"

	certLabels["skiperator.kartverket.no/controller"] = "routing"
	certLabels["skiperator.kartverket.no/source-namespace"] = routing.Namespace

	return certLabels
}

func (r *RoutingReconciler) GetSkiperatorRoutingCertificates(context context.Context) (certmanagerv1.CertificateList, error) {
	certificates := certmanagerv1.CertificateList{}
	err := r.GetClient().List(context, &certificates, client.MatchingLabels{
		"app.kubernetes.io/managed-by":        "skiperator",
		"skiperator.kartverket.no/controller": "routing",
	})

	return certificates, err
}
