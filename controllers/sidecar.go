package controllers

import (
	"context"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	networkingv1beta1api "istio.io/api/networking/v1beta1"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

//+kubebuilder:rbac:groups=skiperator.kartverket.no,resources=applications,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=networking.istio.io,resources=sidecars,verbs=get;list;watch;create;update;patch;delete

type SidecarReconciler struct {
	client client.Client
	scheme *runtime.Scheme
}

func (r *SidecarReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.client = mgr.GetClient()
	r.scheme = mgr.GetScheme()

	return newControllerManagedBy[*skiperatorv1alpha1.Application](mgr).
		Owns(&networkingv1beta1.Sidecar{}).
		Complete(r)
}

func (r *SidecarReconciler) Reconcile(ctx context.Context, application *skiperatorv1alpha1.Application) (reconcile.Result, error) {
	application.FillDefaults()

	sidecar := networkingv1beta1.Sidecar{ObjectMeta: metav1.ObjectMeta{Namespace: application.Namespace, Name: application.Name}}
	_, err := ctrlutil.CreateOrPatch(ctx, r.client, &sidecar, func() error {
		// Set application as owner of the sidecar
		err := ctrlutil.SetControllerReference(application, &sidecar, r.scheme)
		if err != nil {
			return err
		}

		sidecar.Spec.OutboundTrafficPolicy = &networkingv1beta1api.OutboundTrafficPolicy{}
		sidecar.Spec.OutboundTrafficPolicy.Mode = networkingv1beta1api.OutboundTrafficPolicy_REGISTRY_ONLY

		return nil
	})
	return reconcile.Result{}, err
}
