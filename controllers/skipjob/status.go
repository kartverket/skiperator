package applicationcontroller

import (
	"context"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
)

const (
	ConditionRunning  = "Running"
	ConditionFinished = "Finished"
)

func (r *SKIPJobReconciler) GetConditionRunning(skipJob *skiperatorv1alpha1.SKIPJob, status v1.ConditionStatus) v1.Condition {
	return v1.Condition{
		Type:               ConditionRunning,
		Status:             status,
		ObservedGeneration: skipJob.Generation,
		LastTransitionTime: v1.Now(),
		Reason:             "JobStarted",
		Message:            "Job has been created and is now running",
	}
}

func (r *SKIPJobReconciler) GetConditionFinished(skipJob *skiperatorv1alpha1.SKIPJob, status v1.ConditionStatus) v1.Condition {
	return v1.Condition{
		Type:               ConditionFinished,
		Status:             status,
		ObservedGeneration: skipJob.Generation,
		LastTransitionTime: v1.Now(),
		Reason:             "JobFinished",
		Message:            "Job has finished",
	}
}

func (r *SKIPJobReconciler) UpdateStatusWithCondition(ctx context.Context, in *skiperatorv1alpha1.SKIPJob, condition v1.Condition) (*ctrl.Result, error) {
	foundJob := &skiperatorv1alpha1.SKIPJob{}
	err := r.GetClient().Get(ctx, types.NamespacedName{
		Name:      in.Name,
		Namespace: in.Namespace,
	}, foundJob)
	if err != nil {
		return &ctrl.Result{}, err
	}

	if shouldAddCondition(foundJob.Status.Conditions, condition) {
		foundJob.Status.Conditions = append(foundJob.Status.Conditions, condition)

		err = r.GetClient().Status().Update(ctx, foundJob)
		if err != nil {
			return &ctrl.Result{}, err
		}
	}

	return nil, nil
}

func shouldAddCondition(conditions []v1.Condition, conditionToAdd v1.Condition) bool {
	if len(conditions) == 0 {
		return true
	}

	currentCondition := conditions[len(conditions)-1]

	if currentCondition.Status == conditionToAdd.Status && currentCondition.Type == conditionToAdd.Type {
		return false
	}

	return true
}
