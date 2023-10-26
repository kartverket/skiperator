package skipjobcontroller

import (
	"context"
	"fmt"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/util"
	istionetworkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	batchv1 "k8s.io/api/batch/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// +kubebuilder:rbac:groups=skiperator.kartverket.no,resources=skipjobs;skipjobs/status,verbs=get;list;watch;update
// +kubebuilder:rbac:groups=batch,resources=jobs;cronjobs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=pods;pods/ephemeralcontainers,verbs=get;list;watch;create;update;patch;delete

type SKIPJobReconciler struct {
	util.ReconcilerBase
}

func (r *SKIPJobReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		// GenerationChangedPredicate is now only applied to the SkipJob itself to allow status changes on Jobs/CronJobs to affect reconcile loops
		For(&skiperatorv1alpha1.SKIPJob{}, builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		Owns(&batchv1.CronJob{}).
		Owns(&batchv1.Job{}).
		// This is added as the Jobs created by CronJobs are not owned by the SKIPJob directly, but rather through the CronJob
		Watches(&batchv1.Job{}, handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, object client.Object) []reconcile.Request {
			job, isJob := object.(*batchv1.Job)

			if !isJob {
				return nil
			}

			if skipJobName, exists := job.Labels[SKIPJobReferenceLabelKey]; exists {
				return []reconcile.Request{
					{
						types.NamespacedName{
							Namespace: job.Namespace,
							Name:      skipJobName,
						},
					},
				}
			}

			return nil
		})).
		Owns(&networkingv1.NetworkPolicy{}).
		Owns(&istionetworkingv1beta1.ServiceEntry{}).
		// Some NetPol entries are not added unless an application is present. If we reconcile all jobs when there has been changes to NetPols, we can assume
		// that changes to an Applications AccessPolicy will cause a reconciliation of Jobs
		Watches(&networkingv1.NetworkPolicy{}, handler.EnqueueRequestsFromMapFunc(r.getJobsToReconcile)).
		Complete(r)
}

func (r *SKIPJobReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	skipJob := &skiperatorv1alpha1.SKIPJob{}
	err := r.GetClient().Get(ctx, req.NamespacedName, skipJob)

	if errors.IsNotFound(err) {
		return reconcile.Result{}, nil
	} else if err != nil {
		r.EmitWarningEvent(skipJob, "ReconcileStartFail", "something went wrong fetching the SKIPJob, it might have been deleted")
		return reconcile.Result{}, err
	}

	tmpSkipJob := skipJob.DeepCopy()
	err = skipJob.ApplyDefaults()
	if err != nil {
		return reconcile.Result{}, err
	}

	specDiff, err := util.GetObjectDiff(tmpSkipJob.Spec, skipJob.Spec)
	if err != nil {
		return reconcile.Result{}, err
	}
	statusDiff, err := util.GetObjectDiff(tmpSkipJob.Status, skipJob.Status)
	if err != nil {
		return reconcile.Result{}, err
	}

	// If we update the SKIPJob initially on applied defaults before starting reconciling resources we allow all
	// updates to be visible even though the controllerDuties may take some time.
	if len(specDiff) > 0 {
		err := r.GetClient().Update(ctx, skipJob)
		return reconcile.Result{Requeue: true}, err
	}

	if len(statusDiff) > 0 {
		err := r.GetClient().Status().Update(ctx, skipJob)
		return reconcile.Result{Requeue: true}, err
	}

	r.EmitNormalEvent(skipJob, "ReconcileStart", fmt.Sprintf("SKIPJob %v has started reconciliation loop", skipJob.Name))

	controllerDuties := []func(context.Context, *skiperatorv1alpha1.SKIPJob) (reconcile.Result, error){
		r.reconcileServiceAccount,
		r.reconcileNetworkPolicy,
		r.reconcileEgressServiceEntry,
		r.reconcileConfigMap,
		r.reconcileJob,
	}

	for _, fn := range controllerDuties {
		res, err := fn(ctx, skipJob)
		if err != nil {
			return res, err
		} else if res.RequeueAfter > 0 || res.Requeue {
			return res, nil
		}
	}

	r.EmitNormalEvent(skipJob, "ReconcileEnd", fmt.Sprintf("SKIPJob %v has finished reconciliation loop", skipJob.Name))

	err = r.GetClient().Update(ctx, skipJob)
	return reconcile.Result{}, err
}

func (r *SKIPJobReconciler) getJobsToReconcile(ctx context.Context, object client.Object) []reconcile.Request {
	var jobsToReconcile skiperatorv1alpha1.SKIPJobList
	var reconcileRequests []reconcile.Request

	owner := object.GetOwnerReferences()
	if len(owner) == 0 {
		return reconcileRequests
	}

	// Assume only one owner
	if owner[0].Kind != "Application" {
		return reconcileRequests
	}

	err := r.GetClient().List(ctx, &jobsToReconcile)
	if err != nil {
		return nil
	}
	for _, job := range jobsToReconcile.Items {
		reconcileRequests = append(reconcileRequests, reconcile.Request{
			NamespacedName: types.NamespacedName{
				Namespace: job.Namespace,
				Name:      job.Name,
			},
		})
	}
	return reconcileRequests
}
