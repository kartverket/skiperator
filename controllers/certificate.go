package controllers

import (
	"context"
	"fmt"
	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"hash/fnv"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"strings"
)

//+kubebuilder:rbac:groups=skiperator.kartverket.no,resources=applications,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cert-manager.io,resources=certificates,verbs=get;list;watch;create;update;patch;delete

type CertificateReconciler struct {
	client client.Client
	scheme *runtime.Scheme
}

func (r *CertificateReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.client = mgr.GetClient()
	r.scheme = mgr.GetScheme()

	return ctrl.NewControllerManagedBy(mgr).
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

func (r *CertificateReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	// Fetch application and fill defaults
	application := skiperatorv1alpha1.Application{}
	err := r.client.Get(ctx, req.NamespacedName, &application)
	if err != nil {
		err = client.IgnoreNotFound(err)
		return reconcile.Result{}, err
	}
	application.FillDefaults()

	// Keep track of active certificates
	active := make(map[string]struct{}, len(application.Spec.Ingresses))

	// Generate separate gateway for each ingress
	for _, hostname := range application.Spec.Ingresses {
		// Generate certificate name
		hash := fnv.New64()
		_, _ = hash.Write([]byte(hostname))
		name := fmt.Sprintf("%s-%s-ingress-%x", req.Namespace, req.Name, hash.Sum64())
		active[name] = struct{}{}

		certificate := certmanagerv1.Certificate{ObjectMeta: metav1.ObjectMeta{Namespace: "istio-system", Name: name}}
		_, err = ctrlutil.CreateOrPatch(ctx, r.client, &certificate, func() error {
			// Set application as owner of the certificate
			err = ctrlutil.SetControllerReference(&application, &certificate, r.scheme)
			if err != nil {
				return err
			}

			certificate.Spec.IssuerRef.Kind = "ClusterIssuer"
			certificate.Spec.IssuerRef.Name = "lets-encrypt"
			certificate.Spec.DNSNames = []string{hostname}
			certificate.Spec.SecretName = name

			return nil
		})
	}

	// Clear out unused certificates
	certificates := certmanagerv1.CertificateList{}
	err = r.client.List(ctx, &certificates, client.InNamespace("istio-system"))
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
