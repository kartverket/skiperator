package applicationcontroller

import (
	"context"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"time"
)

const (
	ConditionRunning  = "Running"
	ConditionFinished = "Finished"
)

func (r *SKIPJobReconciler) GetConditionRunning(skipJob *skiperatorv1alpha1.SKIPJob) v1.Condition {
	return v1.Condition{
		Type:               ConditionRunning,
		Status:             "True",
		ObservedGeneration: skipJob.Generation,
		LastTransitionTime: v1.Now(),
		Reason:             "JobStarted",
		Message:            "Job has been created and is now starting",
	}
}

func (r *SKIPJobReconciler) GetConditionFinished(skipJob *skiperatorv1alpha1.SKIPJob) v1.Condition {
	return v1.Condition{
		Type:               ConditionFinished,
		Status:             "True",
		ObservedGeneration: skipJob.Generation,
		LastTransitionTime: v1.Now(),
		Reason:             "JobFinished",
		Message:            "Job has finished",
	}
}

// TODO Do not append condition if last condition is same
func (r *SKIPJobReconciler) UpdateStatusWithCondition(ctx context.Context, in *skiperatorv1alpha1.SKIPJob, condition v1.Condition) (*ctrl.Result, error) {
	foundJob := &skiperatorv1alpha1.SKIPJob{}
	err := r.GetClient().Get(ctx, types.NamespacedName{
		Name:      in.Name,
		Namespace: in.Namespace,
	}, foundJob)
	if err != nil {
		// The job may not have been created yet, so requeue
		return &ctrl.Result{RequeueAfter: 5 * time.Second}, err
	}

	foundJob.Status.Conditions = append(foundJob.Status.Conditions, condition)

	err = r.GetClient().Status().Update(ctx, foundJob)
	if err != nil {
		return &ctrl.Result{}, err
	}

	return nil, nil
}
