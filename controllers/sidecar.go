package controllers

import (
	"context"

	util "github.com/kartverket/skiperator/pkg/util"
	networkingv1beta1api "istio.io/api/networking/v1beta1"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

//+kubebuilder:rbac:groups=core,resources=namespaces,verbs=get;list;watch
//+kubebuilder:rbac:groups=networking.istio.io,resources=sidecars,verbs=get;list;watch;create;update;patch;delete

type SidecarReconciler struct {
	client client.Client
	scheme *runtime.Scheme
}

func (r *SidecarReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.client = mgr.GetClient()
	r.scheme = mgr.GetScheme()

	return newControllerManagedBy[*corev1.Namespace](mgr).
		For(&corev1.Namespace{}, builder.WithPredicates(
			matchesPredicate[*corev1.Namespace](util.IsNotExcludedNamespace),
		)).
		Owns(&networkingv1beta1.Sidecar{}).
		Complete(r)
}

func (r *SidecarReconciler) Reconcile(ctx context.Context, namespace *corev1.Namespace) (reconcile.Result, error) {
	sidecar := networkingv1beta1.Sidecar{ObjectMeta: metav1.ObjectMeta{Namespace: namespace.Name, Name: "sidecar"}}
	_, err := ctrlutil.CreateOrPatch(ctx, r.client, &sidecar, func() error {
		// Set namespace as owner of the sidecar
		err := ctrlutil.SetControllerReference(namespace, &sidecar, r.scheme)
		if err != nil {
			return err
		}

		sidecar.Spec.OutboundTrafficPolicy = &networkingv1beta1api.OutboundTrafficPolicy{}
		sidecar.Spec.OutboundTrafficPolicy.Mode = networkingv1beta1api.OutboundTrafficPolicy_REGISTRY_ONLY

		return nil
	})
	return reconcile.Result{}, err
}
