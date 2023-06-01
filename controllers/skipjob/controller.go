package applicationcontroller

import (
	"context"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/util"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// +kubebuilder:rbac:groups=skiperator.kartverket.no,resources=skipjobs;skipjobs/status,verbs=get;list;watch;update
// +kubebuilder:rbac:groups=batch,resources=jobs;cronjobs,verbs=get;list;watch;create;update;patch;delete

type SKIPJobReconciler struct {
	util.ReconcilerBase
}

func (r *SKIPJobReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&skiperatorv1alpha1.SKIPJob{}).
		Owns(&batchv1.CronJob{}).
		Owns(&batchv1.Job{}).
		Owns(&networkingv1.NetworkPolicy{}).
		WithEventFilter(predicate.GenerationChangedPredicate{}).
		Complete(r)
}

func (r *SKIPJobReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	skipJob := &skiperatorv1alpha1.SKIPJob{}
	err := r.GetClient().Get(ctx, req.NamespacedName, skipJob)

	if errors.IsNotFound(err) {
		return reconcile.Result{}, nil
	} else if err != nil {
		r.GetRecorder().Eventf(
			skipJob,
			corev1.EventTypeNormal, "ReconcileStartFail",
			"Something went wrong fetching the SKIPJob. It might have been deleted",
		)
		return reconcile.Result{}, err
	}

	r.GetRecorder().Eventf(
		skipJob,
		corev1.EventTypeNormal, "ReconcileStart",
		"SKIPJob "+skipJob.Name+" has started reconciliation loop",
	)

	controllerDuties := []func(context.Context, *skiperatorv1alpha1.SKIPJob) (reconcile.Result, error){
		r.reconcileJob,
		r.reconcileServiceAccount,
		r.reconcileNetworkPolicy,
	}

	for _, fn := range controllerDuties {
		if _, err := fn(ctx, skipJob); err != nil {
			return reconcile.Result{}, err
		}
	}

	r.GetRecorder().Eventf(
		skipJob,
		corev1.EventTypeNormal, "ReconcileEnd",
		"SKIPJob "+skipJob.Name+" has finished reconciliation loop",
	)

	return reconcile.Result{}, err
}
