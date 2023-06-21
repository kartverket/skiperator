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

func (r *SKIPJobReconciler) UpdateStatusWithCondition(ctx context.Context, in *skiperatorv1alpha1.SKIPJob, conditions []v1.Condition) (*ctrl.Result, error) {
	foundJob := &skiperatorv1alpha1.SKIPJob{}
	err := r.GetClient().Get(ctx, types.NamespacedName{
		Name:      in.Name,
		Namespace: in.Namespace,
	}, foundJob)
	if err != nil {
		return &ctrl.Result{}, err
	}

	for _, conditionToAdd := range conditions {
		currentCondition, isNotEmpty := r.GetLastCondition(foundJob.Status.Conditions)

		isSameType := conditionsHaveSameType(currentCondition, &conditionToAdd)
		isSameStatus := conditionsHaveSameStatus(currentCondition, &conditionToAdd)

		if !isNotEmpty || !isSameType {
			foundJob.Status.Conditions = append(foundJob.Status.Conditions, conditionToAdd)
		}

		if isSameType && isSameStatus {
			continue
		}

		if isSameType && !isSameStatus {
			*currentCondition = conditionToAdd
		}
	}

	err = r.GetClient().Status().Update(ctx, foundJob)
	if err != nil {
		return &ctrl.Result{}, err
	}

	return nil, nil
}

func (r *SKIPJobReconciler) GetLastCondition(conditions []v1.Condition) (*v1.Condition, bool) {
	if len(conditions) == 0 {
		return &v1.Condition{}, false
	}

	return &conditions[len(conditions)-1], true
}

func conditionsHaveSameStatus(condition1 *v1.Condition, condition2 *v1.Condition) bool {
	return condition1.Status == condition2.Status
}

func conditionsHaveSameType(condition1 *v1.Condition, condition2 *v1.Condition) bool {
	return condition1.Type == condition2.Type
}
