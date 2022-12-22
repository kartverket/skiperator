package applicationcontroller

import (
	"context"
	"fmt"
	"hash/fnv"

	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func IsSkiperatorOwnedCertificate(certificate *certmanagerv1.Certificate) bool {
	match := certificate.Labels["app.kubernetes.io/managed-by"] == "skiperator"
	return match
}

func (r *ApplicationReconciler) reconcileCertificate(ctx context.Context, application *skiperatorv1alpha1.Application) (reconcile.Result, error) {
	controllerName := "Certificate"
	r.SetControllerProgressing(ctx, application, controllerName)

	// Generate separate gateway for each ingress
	for _, hostname := range application.Spec.Ingresses {
		// Generate certificate name
		hash := fnv.New64()
		_, _ = hash.Write([]byte(hostname))
		name := fmt.Sprintf("%s-%s-ingress-%x", application.Namespace, application.Name, hash.Sum64())

		certificate := certmanagerv1.Certificate{ObjectMeta: metav1.ObjectMeta{Namespace: "istio-system", Name: name}}
		_, err := ctrlutil.CreateOrPatch(ctx, r.GetClient(), &certificate, func() error {
			certificate.Spec.IssuerRef.Kind = "ClusterIssuer"
			certificate.Spec.IssuerRef.Name = "cluster-issuer" // Name defined in https://github.com/kartverket/certificate-management/blob/main/clusterissuer.tf
			certificate.Spec.DNSNames = []string{hostname}
			certificate.Spec.SecretName = name

			certLabels := certificate.Labels
			if len(certLabels) == 0 {
				certLabels = make(map[string]string)
			}
			certLabels["app.kubernetes.io/managed-by"] = "skiperator"

			certificate.Labels = certLabels

			return nil
		})
		if err != nil {
			r.SetControllerError(ctx, application, controllerName, err)
			return reconcile.Result{}, err
		}
	}

	r.SetControllerFinishedOutcome(ctx, application, controllerName, nil)

	return reconcile.Result{}, nil
}
