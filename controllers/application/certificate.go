package applicationcontroller

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	util "github.com/kartverket/skiperator/pkg/util"
	"golang.org/x/exp/slices"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *ApplicationReconciler) ApplicationFromCertificate(obj client.Object) []reconcile.Request {
	cert := obj.(*certmanagerv1.Certificate)

	if cert.Namespace != "istio-system" {
		return nil
	}

	segments := strings.SplitN(cert.Name, "-", 4)
	if len(segments) != 4 {
		return nil
	}

	return []reconcile.Request{
		{NamespacedName: types.NamespacedName{Namespace: segments[0], Name: segments[1]}},
	}
}

func (r *ApplicationReconciler) reconcileCertificate(ctx context.Context, application *skiperatorv1alpha1.Application) (reconcile.Result, error) {
	controllerName := "Certificate"
	r.SetControllerProgressing(ctx, application, controllerName)

	// Generate separate gateway for each ingress
	for _, hostname := range application.Spec.Ingresses {
		name := fmt.Sprintf("%s-%s-ingress-%x", application.Namespace, application.Name, util.GenerateHashFromName(hostname))

		certificate := certmanagerv1.Certificate{ObjectMeta: metav1.ObjectMeta{Namespace: "istio-system", Name: name}}
		_, err := ctrlutil.CreateOrPatch(ctx, r.GetClient(), &certificate, func() error {
			certificate.Spec.IssuerRef.Kind = "ClusterIssuer"
			certificate.Spec.IssuerRef.Name = "cluster-issuer" // Name defined in https://github.com/kartverket/certificate-management/blob/main/clusterissuer.tf
			certificate.Spec.DNSNames = []string{hostname}
			certificate.Spec.SecretName = name

			return nil
		})
		if err != nil {
			r.SetControllerError(ctx, application, controllerName, err)
			return reconcile.Result{}, err
		}
	}

	// Clear out unused certs
	certificates := certmanagerv1.CertificateList{}
	err := r.GetClient().List(ctx, &certificates, client.InNamespace("istio-system"))
	if err != nil {
		r.SetControllerError(ctx, application, controllerName, err)
		return reconcile.Result{}, err
	}

	for _, certificate := range certificates.Items {

		certificateInApplicationSpecIndex := slices.IndexFunc(application.Spec.Ingresses, func(hostname string) bool {
			certificateName := fmt.Sprintf("%s-%s-ingress-%x", application.Namespace, application.Name, util.GenerateHashFromName(hostname))
			return certificate.Name == certificateName
		})
		certificateInApplicationSpec := certificateInApplicationSpecIndex != -1
		if certificateInApplicationSpec {
			continue
		}

		// We want to delete certificate which are not in the spec, but still "owned" by the application.
		// This should be the case for any certificate not in spec from the earlier continue, if the name still matches <namespace>-<application-name>-ingress-*
		applicationNamespacedName := application.Namespace + "-" + application.Name
		certNameMatchesApplicationNamespacedName, _ := regexp.MatchString("^"+applicationNamespacedName+"-ingress-.+$", certificate.Name)
		if !certNameMatchesApplicationNamespacedName {
			continue
		}

		// Delete the rest
		err = r.GetClient().Delete(ctx, &certificate)
		err = client.IgnoreNotFound(err)
		if err != nil {
			r.SetControllerError(ctx, application, controllerName, err)
			return reconcile.Result{}, err
		}
	}

	r.SetControllerFinishedOutcome(ctx, application, controllerName, err)

	return reconcile.Result{}, err
}
