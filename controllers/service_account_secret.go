package controllers

import (
	"context"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

//+kubebuilder:rbac:groups=skiperator.kartverket.no,resources=applications,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=secret,verbs=get;list;watch;create;update;patch;delete

type ServiceAccountSecretReconciler struct {
	client client.Client
	scheme *runtime.Scheme
}

func (r *ServiceAccountSecretReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.client = mgr.GetClient()
	r.scheme = mgr.GetScheme()

	return ctrl.NewControllerManagedBy(mgr).
		For(&skiperatorv1alpha1.Application{}).
		Owns(&corev1.Secret{}).
		Complete(r)
}

func (r *ServiceAccountSecretReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	// Fetch application and fill defaults
	application := skiperatorv1alpha1.Application{}
	err := r.client.Get(ctx, req.NamespacedName, &application)
	if err != nil {
		err = client.IgnoreNotFound(err)
		return reconcile.Result{}, err
	}
	application.FillDefaults()

	secret := corev1.Secret{ObjectMeta: metav1.ObjectMeta{Namespace: req.Namespace, Name: req.Name + "-token"}}
	_, err = ctrlutil.CreateOrPatch(ctx, r.client, &secret, func() error {
		// Set application as owner of the sidecar
		err = ctrlutil.SetControllerReference(&application, &secret, r.scheme)
		if err != nil {
			return err
		}

		secret.Type = corev1.SecretTypeServiceAccountToken

		secret.Annotations = make(map[string]string, 1)
		secret.Annotations[corev1.ServiceAccountNameKey] = application.Name

		return nil
	})
	return reconcile.Result{}, err
}
