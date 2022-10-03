package controllers

import (
	"context"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	securityv1beta1api "istio.io/api/security/v1beta1"
	typev1beta1 "istio.io/api/type/v1beta1"
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

//+kubebuilder:rbac:groups=skiperator.kartverket.no,resources=applications,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=security.istio.io,resources=peerauthentications,verbs=get;list;watch;create;update;patch;delete

type PeerAuthenticationReconciler struct {
	client client.Client
	scheme *runtime.Scheme
}

func (r *PeerAuthenticationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.client = mgr.GetClient()
	r.scheme = mgr.GetScheme()

	return newControllerManagedBy[*skiperatorv1alpha1.Application](mgr).
		Owns(&securityv1beta1.PeerAuthentication{}).
		Complete(r)
}

func (r *PeerAuthenticationReconciler) Reconcile(ctx context.Context, application *skiperatorv1alpha1.Application) (reconcile.Result, error) {
	application.FillDefaults()

	peerAuthentication := securityv1beta1.PeerAuthentication{ObjectMeta: metav1.ObjectMeta{Namespace: application.Namespace, Name: application.Name}}
	_, err := ctrlutil.CreateOrPatch(ctx, r.client, &peerAuthentication, func() error {
		// Set application as owner of the peer authentication
		err := ctrlutil.SetControllerReference(application, &peerAuthentication, r.scheme)
		if err != nil {
			return err
		}

		peerAuthentication.Spec.Selector = &typev1beta1.WorkloadSelector{}
		labels := map[string]string{"app": application.Name}
		peerAuthentication.Spec.Selector.MatchLabels = labels

		peerAuthentication.Spec.Mtls = &securityv1beta1api.PeerAuthentication_MutualTLS{}
		peerAuthentication.Spec.Mtls.Mode = securityv1beta1api.PeerAuthentication_MutualTLS_STRICT

		return nil
	})
	return reconcile.Result{}, err
}
