package controllers

import (
	"context"
	"fmt"

	"github.com/kartverket/skiperator/pkg/resourceprocessor"
	"github.com/kartverket/skiperator/pkg/resourceschemas"

	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	skiperatorv1beta1 "github.com/kartverket/skiperator/api/v1beta1"
	"github.com/kartverket/skiperator/internal/controllers/common"
	"github.com/kartverket/skiperator/pkg/log"
	. "github.com/kartverket/skiperator/pkg/reconciliation"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/gcp/auth"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/istio/serviceentry"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/istio/telemetry"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/job"
	networkpolicy "github.com/kartverket/skiperator/pkg/resourcegenerator/networkpolicy/dynamic"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/prometheus"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/resourceutils"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/serviceaccount"
	istionetworkingv1 "istio.io/client-go/pkg/apis/networking/v1"
	telemetryv1 "istio.io/client-go/pkg/apis/telemetry/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	ConditionRunning  = "Running"
	ConditionFinished = "Finished"
	ConditionFailed   = "Failed"

	skipJobFinalizer = "skip.statkart.no/finalizer"
)

// +kubebuilder:rbac:groups=skiperator.kartverket.no,resources=skipjobs;skipjobs/status,verbs=get;list;watch;update
// +kubebuilder:rbac:groups=batch,resources=jobs;cronjobs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=telemetry.istio.io,resources=telemetries,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=pods;pods/ephemeralcontainers,verbs=get;list;watch;create;update;patch;delete

// leave an empty line over this comment
type SKIPJobReconciler struct {
	common.ReconcilerBase
	client.Client
	Scheme *runtime.Scheme
}

// TODO Watch applications that are using dynamic port allocation
func (r *SKIPJobReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		// GenerationChangedPredicate is now only applied to the SkipJob itself to allow status changes on Jobs/CronJobs to affect reconcile loops
		For(&skiperatorv1beta1.SKIPJob{}, builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		Owns(&batchv1.CronJob{}).
		Owns(&batchv1.Job{}).
		// This is added as the Jobs created by CronJobs are not owned by the SKIPJob directly, but rather through the CronJob
		Watches(&batchv1.Job{}, handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, object client.Object) []reconcile.Request {
			batchJob, isJob := object.(*batchv1.Job)

			if !isJob {
				return nil
			}

			if skipJobName, exists := batchJob.Labels[skiperatorv1beta1.SKIPJobReferenceLabelKey]; exists {
				return []reconcile.Request{
					{
						types.NamespacedName{
							Namespace: batchJob.Namespace,
							Name:      skipJobName,
						},
					},
				}
			}

			return nil
		})).
		Owns(&networkingv1.NetworkPolicy{}).
		Owns(&istionetworkingv1.ServiceEntry{}).
		Owns(&telemetryv1.Telemetry{}).
		// Some NetPol entries are not added unless an application is present. If we reconcile all jobs when there has been changes to NetPols, we can assume
		// that changes to an Applications AccessPolicy will cause a reconciliation of Jobs
		Watches(&networkingv1.NetworkPolicy{}, handler.EnqueueRequestsFromMapFunc(r.getJobsToReconcile)).
		Complete(r)
}

func (r *SKIPJobReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	rLog := log.NewLogger().WithName(fmt.Sprintf("skipjob-controller: %s", req.Name))
	rLog.Debug("Starting reconcile for request", "request", req.Name)

	skipJob, err := r.getSKIPJob(ctx, req)
	if skipJob == nil {
		return common.DoNotRequeue()
	} else if err != nil {
		r.EmitWarningEvent(skipJob, "ReconcileStartFail", "something went wrong fetching the SKIPJob, it might have been deleted")
		return common.RequeueWithError(err)
	}

	isSkipJobMarkedToBeDeleted := skipJob.GetDeletionTimestamp() != nil
	if isSkipJobMarkedToBeDeleted {
		if err = r.finalizeSkipJob(skipJob, ctx); err != nil {
			return ctrl.Result{}, err
		}
		return common.DoNotRequeue()
	}

	tmpSkipJob := skipJob.DeepCopy()
	//TODO make sure we don't update the skipjob/application/routing after this step, it will cause endless reconciliations
	//check that resource request limit 0.3 doesn't overwrite to 300m
	err = r.setSKIPJobDefaults(ctx, skipJob)
	if err != nil {
		return common.RequeueWithError(err)
	}

	specDiff, err := common.GetObjectDiff(tmpSkipJob.Spec, skipJob.Spec)
	if err != nil {
		return common.RequeueWithError(err)
	}

	statusDiff, err := common.GetObjectDiff(tmpSkipJob.Status, skipJob.Status)
	if err != nil {
		return common.RequeueWithError(err)
	}

	if len(statusDiff) > 0 {
		rLog.Debug("Status has changed", "diff", statusDiff)
		err = r.GetClient().Status().Update(ctx, skipJob)
		return reconcile.Result{Requeue: true}, err
	}

	// Finalizer check is due to a bug when updating using controller-runtime
	// If we update the SKIPJob initially on applied defaults before starting reconciling resources we allow all
	// updates to be visible even though the controllerDuties may take some time.
	if len(specDiff) > 0 || (!ctrlutil.ContainsFinalizer(tmpSkipJob, skipJobFinalizer) && ctrlutil.ContainsFinalizer(skipJob, skipJobFinalizer)) {
		rLog.Debug("Queuing for spec diff")
		err = r.GetClient().Update(ctx, skipJob)
		return reconcile.Result{Requeue: true}, err
	}

	//We try to feed the access policy with port values dynamically,
	//if unsuccessfull we just don't set ports, and rely on podselectors
	r.UpdateAccessPolicy(ctx, skipJob)

	//Start the actual reconciliation
	rLog.Debug("Starting reconciliation loop")
	r.SetProgressingState(ctx, skipJob, fmt.Sprintf("SKIPJob %v has started reconciliation loop", skipJob.Name))

	istioEnabled := r.IsIstioEnabledForNamespace(ctx, skipJob.Namespace)
	identityConfigMap, err := r.GetIdentityConfigMap(ctx)
	if err != nil {
		rLog.Error(err, "can't find identity config map")
	} //TODO Error state?

	reconciliation := NewJobReconciliation(ctx, skipJob, rLog, istioEnabled, r.GetRestConfig(), identityConfigMap)

	resourceGeneration := []reconciliationFunc{
		serviceaccount.Generate,
		networkpolicy.Generate,
		serviceentry.Generate,
		auth.Generate,
		job.Generate,
		prometheus.Generate,
		telemetry.Generate,
	}

	for _, f := range resourceGeneration {
		if err := f(reconciliation); err != nil {
			rLog.Error(err, "failed to generate skipjob resource")
			// At this point we don't have the gvk of the resource yet, so we can't set subresource status.
			r.SetErrorState(ctx, skipJob, err, "failed to generate skipjob resource", "ResourceGenerationFailure")
			return common.RequeueWithError(err)
		}
	}

	if err = r.setResourceDefaults(reconciliation.GetResources(), skipJob); err != nil {
		rLog.Error(err, "error when trying to set resource defaults")
		r.SetErrorState(ctx, skipJob, err, "failed to set skipjob resource defaults", "ResourceDefaultsFailure")
		return common.RequeueWithError(err)
	}

	processor := resourceprocessor.NewResourceProcessor(r.GetClient(), resourceschemas.GetJobSchemas(r.GetScheme()), r.GetScheme())

	if errs := processor.Process(reconciliation); len(errs) > 0 {
		for _, err = range errs {
			rLog.Error(err, "failed to process resource")
			r.EmitWarningEvent(skipJob, "ReconcileEndFail", fmt.Sprintf("Failed to process skipjob resources: %s", err.Error()))
		}
		r.SetErrorState(ctx, skipJob, fmt.Errorf("found %d errors", len(errs)), "failed to process skipjob resources, see subresource status", "ProcessorFailure")
		return common.RequeueWithError(err)
	}

	//TODO consider if we need better handling of status updates in context of summary, conditions and subresources
	if err = r.updateConditions(ctx, skipJob); err != nil {
		rLog.Error(err, "failed to update conditions")
		r.SetErrorState(ctx, skipJob, err, "failed to update conditions", "ConditionsFailure")
		return common.RequeueWithError(err)
	}

	r.SetSyncedState(ctx, skipJob, "SKIPJob has been reconciled")

	return common.RequeueWithError(err)
}

func (r *SKIPJobReconciler) getSKIPJob(ctx context.Context, req reconcile.Request) (*skiperatorv1beta1.SKIPJob, error) {
	skipJob := &skiperatorv1beta1.SKIPJob{}
	if err := r.GetClient().Get(ctx, req.NamespacedName, skipJob); err != nil {
		if errors.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("error when trying to get routing: %w", err)
	}

	return skipJob, nil
}

func (r *SKIPJobReconciler) setSKIPJobDefaults(ctx context.Context, skipJob *skiperatorv1beta1.SKIPJob) error {
	skipJob.FillDefaultSpec()

	if !ctrlutil.ContainsFinalizer(skipJob, skipJobFinalizer) {
		ctrlutil.AddFinalizer(skipJob, skipJobFinalizer)
	}

	resourceutils.SetSKIPJobLabels(skipJob, skipJob)
	skipJob.FillDefaultStatus()

	return nil
}

func (r *SKIPJobReconciler) setResourceDefaults(resources []client.Object, skipJob *skiperatorv1beta1.SKIPJob) error {
	for _, resource := range resources {
		if err := r.SetSubresourceDefaults(resources, skipJob); err != nil {
			return err
		}
		resourceutils.SetSKIPJobLabels(resource, skipJob)
	}
	return nil
}

func (r *SKIPJobReconciler) finalizeSkipJob(skipJob *skiperatorv1beta1.SKIPJob, ctx context.Context) error {
	if ctrlutil.ContainsFinalizer(skipJob, skipJobFinalizer) {
		ctrlutil.RemoveFinalizer(skipJob, skipJobFinalizer)
		err := r.GetClient().Update(ctx, skipJob)
		if err != nil {
			return fmt.Errorf("something went wrong when trying to finalize SKIPJob: %w", err)
		}
	}
	return nil
}

func (r *SKIPJobReconciler) getJobsToReconcile(ctx context.Context, object client.Object) []reconcile.Request {
	var jobsToReconcile skiperatorv1beta1.SKIPJobList
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
	for _, j := range jobsToReconcile.Items {
		reconcileRequests = append(reconcileRequests, reconcile.Request{
			NamespacedName: types.NamespacedName{
				Namespace: j.Namespace,
				Name:      j.Name,
			},
		})
	}
	return reconcileRequests
}

func (r *SKIPJobReconciler) getConditionRunning(skipJob *skiperatorv1beta1.SKIPJob, status v1.ConditionStatus) v1.Condition {
	return v1.Condition{
		Type:               ConditionRunning,
		Status:             status,
		ObservedGeneration: skipJob.Generation,
		LastTransitionTime: v1.Now(),
		Reason:             "JobRunning",
		Message:            "Job has been created and is now running",
	}
}

func (r *SKIPJobReconciler) getConditionFinished(skipJob *skiperatorv1beta1.SKIPJob, status v1.ConditionStatus) v1.Condition {
	return v1.Condition{
		Type:               ConditionFinished,
		Status:             status,
		ObservedGeneration: skipJob.Generation,
		LastTransitionTime: v1.Now(),
		Reason:             "JobFinished",
		Message:            "Job has finished",
	}
}

func (r *SKIPJobReconciler) getConditionFailed(skipJob *skiperatorv1beta1.SKIPJob, status v1.ConditionStatus, err *string) v1.Condition {
	conditionMessage := "Job failed previous run"
	if err != nil {
		conditionMessage = fmt.Sprintf("%v: %v", conditionMessage, *err)
	}
	return v1.Condition{
		Type:               ConditionFailed,
		Status:             status,
		ObservedGeneration: skipJob.Generation,
		LastTransitionTime: v1.Now(),
		Reason:             "JobFailed",
		Message:            conditionMessage,
	}
}

func (r *SKIPJobReconciler) updateConditions(ctx context.Context, skipJob *skiperatorv1beta1.SKIPJob) error {
	jobList := &batchv1.JobList{}
	err := r.GetClient().List(ctx, jobList,
		client.InNamespace(skipJob.Namespace),
		client.MatchingLabels(skipJob.GetDefaultLabels()),
	)
	if err != nil {
		return fmt.Errorf("failed to list jobs: %w", err)
	}

	//find last job to set conditions, cronjobs have multiple jobs
	lastJob := &batchv1.Job{}
	for _, liveJob := range jobList.Items {
		if lastJob.CreationTimestamp.Before(&liveJob.CreationTimestamp) {
			lastJob = &liveJob
		}
	}

	if len(jobList.Items) == 0 {
		skipJob.Status.Conditions = []v1.Condition{
			r.getConditionFailed(skipJob, v1.ConditionFalse, nil),
			r.getConditionRunning(skipJob, v1.ConditionFalse),
			r.getConditionFinished(skipJob, v1.ConditionFalse),
		}
	} else if isFailed, failedJobMessage := isFailedJob(lastJob); isFailed {
		skipJob.Status.Conditions = []v1.Condition{
			r.getConditionFailed(skipJob, v1.ConditionTrue, &failedJobMessage),
			r.getConditionRunning(skipJob, v1.ConditionFalse),
			r.getConditionFinished(skipJob, v1.ConditionFalse),
		}
	} else if lastJob.Status.CompletionTime != nil {
		skipJob.Status.Conditions = []v1.Condition{
			r.getConditionFailed(skipJob, v1.ConditionFalse, nil),
			r.getConditionRunning(skipJob, v1.ConditionFalse),
			r.getConditionFinished(skipJob, v1.ConditionTrue),
		}
	} else {
		skipJob.Status.Conditions = []v1.Condition{
			r.getConditionFailed(skipJob, v1.ConditionFalse, nil),
			r.getConditionRunning(skipJob, v1.ConditionTrue),
			r.getConditionFinished(skipJob, v1.ConditionFalse),
		}
	}

	// Invalid port condition
	accessPolicy := skipJob.Spec.AccessPolicy

	if accessPolicy != nil && !common.IsInternalRulesValid(accessPolicy) {
		skipJob.Status.Conditions = append(skipJob.Status.Conditions, common.GetInternalRulesCondition(skipJob, v1.ConditionFalse))
		skipJob.Status.AccessPolicies = skiperatorv1beta1.INVALIDCONFIG
	} else if accessPolicy != nil && !common.IsExternalRulesValid(accessPolicy) {
		skipJob.Status.Conditions = append(skipJob.Status.Conditions, common.GetExternalRulesCondition(skipJob, v1.ConditionFalse))
		skipJob.Status.AccessPolicies = skiperatorv1beta1.INVALIDCONFIG
	} else {
		skipJob.Status.Conditions = append(skipJob.Status.Conditions, common.GetInternalRulesCondition(skipJob, v1.ConditionTrue))
		skipJob.Status.Conditions = append(skipJob.Status.Conditions, common.GetExternalRulesCondition(skipJob, v1.ConditionTrue))
		skipJob.Status.AccessPolicies = skiperatorv1beta1.READY
	}

	return nil
}

// think it can be done easier
func isFailedJob(job *batchv1.Job) (bool, string) {
	for _, condition := range job.Status.Conditions {
		if condition.Type == ConditionFailed && condition.Status == corev1.ConditionTrue {
			return true, condition.Message
		}
	}
	return false, ""
}
