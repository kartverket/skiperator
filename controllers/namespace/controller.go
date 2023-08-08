package namespacecontroller

import (
	"context"
	"fmt"
	"github.com/kartverket/skiperator/pkg/util"
	istionetworkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type NamespaceReconciler struct {
	util.ReconcilerBase
	Token    string
	Registry string
}

func (r *NamespaceReconciler) isExcludedNamespace(ctx context.Context, namespace string) bool {
	configMapNamespacedName := types.NamespacedName{Namespace: "skiperator-system", Name: "namespace-exclusions"}

	namespaceExclusionCMap, err := util.GetConfigMap(r.GetClient(), ctx, configMapNamespacedName)
	if err != nil {
		util.ErrDoPanic(err, "Something went wrong getting namespace-exclusion config map: %v")
	}

	nameSpacesToExclude := namespaceExclusionCMap.Data

	exclusion, keyExists := nameSpacesToExclude[namespace]

	return (keyExists && exclusion == "true")
}

//+kubebuilder:rbac:groups=core,resources=namespaces,verbs=get;list;watch
//+kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=networking.istio.io,resources=sidecars,verbs=get;list;watch;create;update;patch;delete

func (r *NamespaceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Namespace{}).
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
		r.EmitWarningEvent(namespace, "ReconcileStartFail", "something went wrong fetching the namespace, it might have been deleted")

		return reconcile.Result{}, err
	}

	if r.isExcludedNamespace(ctx, namespace.Name) {
		return reconcile.Result{}, err
	}

	r.EmitNormalEvent(namespace, "ReconcileStart", fmt.Sprintf("Namespace %v has started reconciliation loop", namespace.Name))

	controllerDuties := []func(context.Context, *corev1.Namespace) (reconcile.Result, error){
		r.reconcileDefaultDenyNetworkPolicy,
		r.reconcileImagePullSecret,
		r.reconcileSidecar,
	}

	for _, fn := range controllerDuties {
		if _, err := fn(ctx, namespace); err != nil {
			return reconcile.Result{}, err
		}
	}

	r.EmitNormalEvent(namespace, "ReconcileEnd", fmt.Sprintf("Namespace %v has finished reconciliation loop", namespace.Name))

	return reconcile.Result{}, err
}
