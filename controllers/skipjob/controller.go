package applicationcontroller

import (
	"context"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/util"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// +kubebuilder:rbac:groups=skiperator.kartverket.no,resources=skipjobs;skipjobs/status,verbs=get;list;watch;update

type SKIPJobReconciler struct {
	util.ReconcilerBase
}

func (r *SKIPJobReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&skiperatorv1alpha1.SKIPJob{}).
		WithEventFilter(predicate.GenerationChangedPredicate{}).
		Complete(r)
}

func (r *SKIPJobReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	skipJob := &skiperatorv1alpha1.SKIPJob{}
	err := r.GetClient().Get(ctx, req.NamespacedName, skipJob)

	controllerDuties := []func(context.Context, *skiperatorv1alpha1.SKIPJob) (reconcile.Result, error){}

	for _, fn := range controllerDuties {
		if _, err := fn(ctx, skipJob); err != nil {
			return reconcile.Result{}, err
		}
	}

	return reconcile.Result{}, err
}
