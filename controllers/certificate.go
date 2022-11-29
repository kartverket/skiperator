package controllers

import (
	"context"
	"fmt"
	"hash/fnv"
	"strings"

	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

//+kubebuilder:rbac:groups=skiperator.kartverket.no,resources=applications,verbs=get;list;watch
//+kubebuilder:rbac:groups=cert-manager.io,resources=certificates,verbs=get;list;watch;create;update;patch;delete

type CertificateReconciler struct {
	client client.Client
	scheme *runtime.Scheme
}

func (r *CertificateReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.client = mgr.GetClient()
	r.scheme = mgr.GetScheme()

	return newControllerManagedBy[*skiperatorv1alpha1.Application](mgr).
		For(&skiperatorv1alpha1.Application{}).
		Watches(
			&source.Kind{Type: &certmanagerv1.Certificate{}},
			handler.EnqueueRequestsFromMapFunc(r.applicationFromCertificate),
		).
		Complete(r)
}

func (r *CertificateReconciler) applicationFromCertificate(obj client.Object) []reconcile.Request {
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

func (r *CertificateReconciler) Reconcile(ctx context.Context, application *skiperatorv1alpha1.Application) (reconcile.Result, error) {
	application.FillDefaults()

	// Keep track of active certificates
	active := make(map[string]struct{}, len(application.Spec.Ingresses))

	// Generate separate gateway for each ingress
	for _, hostname := range application.Spec.Ingresses {
		// Generate certificate name
		hash := fnv.New64()
		_, _ = hash.Write([]byte(hostname))
		name := fmt.Sprintf("%s-%s-ingress-%x", application.Namespace, application.Name, hash.Sum64())
		active[name] = struct{}{}

		certificate := certmanagerv1.Certificate{ObjectMeta: metav1.ObjectMeta{Namespace: "istio-system", Name: name}}
		_, err := ctrlutil.CreateOrPatch(ctx, r.client, &certificate, func() error {
			certificate.Spec.IssuerRef.Kind = "ClusterIssuer"
			certificate.Spec.IssuerRef.Name = "cluster-issuer" // Name defined in https://github.com/kartverket/certificate-management/blob/main/clusterissuer.tf
			certificate.Spec.DNSNames = []string{hostname}
			certificate.Spec.SecretName = name

			return nil
		})
		if err != nil {
			return reconcile.Result{}, err
		}
	}

	// Clear out unused certificates
	certificates := certmanagerv1.CertificateList{}
	err := r.client.List(ctx, &certificates, client.InNamespace("istio-system"))
	if err != nil {
		return reconcile.Result{}, err
	}

	for i := range certificates.Items {
		certificate := &certificates.Items[i]

		// Skip unrelated certificates
		segments := strings.SplitN(certificate.Name, "-", 4)
		if len(segments) != 4 || segments[0] != application.Namespace || segments[1] != application.Name {
			continue
		}

		// Skip active certificates
		_, ok := active[certificate.Name]
		if ok {
			continue
		}

		// Delete the rest
		err = r.client.Delete(ctx, certificate)
		err = client.IgnoreNotFound(err)
		if err != nil {
			return reconcile.Result{}, err
		}
	}

	return reconcile.Result{}, err
}
