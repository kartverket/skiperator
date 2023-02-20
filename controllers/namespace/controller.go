package namespacecontroller

import (
	"context"

	"github.com/kartverket/skiperator/pkg/util"
	istionetworkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type NamespaceReconciler struct {
	util.ReconcilerBase
	Token    string
	Registry string
}

//+kubebuilder:rbac:groups=core,resources=namespaces,verbs=get;list;watch
//+kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=networking.istio.io,resources=sidecars,verbs=get;list;watch;create;update;patch;delete

func (r *NamespaceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Namespace{}, builder.WithPredicates(
			util.MatchesPredicate[*corev1.Namespace](util.IsNotExcludedNamespace),
		)).
		Owns(&networkingv1.NetworkPolicy{}).
		Owns(&istionetworkingv1beta1.Sidecar{}).
		Owns(&corev1.Secret{}, builder.WithPredicates(
			util.MatchesPredicate[*corev1.Secret](isImagePullSecret),
		)).
		Complete(r)
}

func (r *NamespaceReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	namespace := &corev1.Namespace{}
	err := r.GetClient().Get(ctx, req.NamespacedName, namespace)

	if errors.IsNotFound(err) {
		return reconcile.Result{}, nil
	} else if err != nil {
		r.GetRecorder().Eventf(
			namespace,
			corev1.EventTypeNormal, "ReconcileStartFail",
			"Something went wrong fetching the namespace. It might have been deleted",
		)
		return reconcile.Result{}, err
	}

	r.GetRecorder().Eventf(
		namespace,
		corev1.EventTypeNormal, "ReconcileStart",
		"Namespace "+namespace.Name+" has started reconciliation loop",
	)

	_, err = r.reconcileDefaultDenyNetworkPolicy(ctx, namespace)
	if err != nil {
		return reconcile.Result{}, err
	}

	_, err = r.reconcileImagePullSecret(ctx, namespace)
	if err != nil {
		return reconcile.Result{}, err
	}

	_, err = r.reconcileSidecar(ctx, namespace)
	if err != nil {
		return reconcile.Result{}, err
	}

	r.GetRecorder().Eventf(
		namespace,
		corev1.EventTypeNormal, "ReconcileEnd",
		"Namespace "+namespace.Name+" has finished reconciliation loop",
	)

	return reconcile.Result{}, err
}
