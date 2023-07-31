package skipjobcontroller

import (
	"context"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"
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

func (r *SKIPJobReconciler) UpdateStatusWithCondition(ctx context.Context, in *skiperatorv1alpha1.SKIPJob, conditions []v1.Condition) error {
	jobConditions := in.Status.Conditions

	for _, conditionToAdd := range conditions {
		currentCondition, exists := r.GetSameConditionIfExists(&jobConditions, &conditionToAdd)
		if !exists {
			in.Status.Conditions = append(in.Status.Conditions, conditionToAdd)
			continue
		}

		isSameType := conditionsHaveSameType(currentCondition, &conditionToAdd)
		isSameStatus := conditionsHaveSameStatus(currentCondition, &conditionToAdd)

		if isSameType && isSameStatus {
			continue
		}

		if isSameType && !isSameStatus {
			r.deleteCondition(ctx, in, *currentCondition)
			in.Status.Conditions = append(in.Status.Conditions, conditionToAdd)
		}
	}

	err := r.GetClient().Status().Update(ctx, in)
	if err != nil {
		return err
	}

	return nil
}

func (r *SKIPJobReconciler) deleteCondition(ctx context.Context, skipJob *skiperatorv1alpha1.SKIPJob, conditionToDelete v1.Condition) error {
	log := log.FromContext(ctx)
	var newConditions []v1.Condition

	for _, condition := range skipJob.Status.Conditions {
		if condition.Type != conditionToDelete.Type {
			newConditions = append(newConditions, condition)
		}
	}

	skipJob.Status.Conditions = newConditions
	err := r.GetClient().Status().Update(ctx, skipJob)
	if err != nil {
		log.Error(err, "skipjob could not delete condition")
		return err
	}
	return nil
}

func (r *SKIPJobReconciler) GetLastCondition(conditions []v1.Condition) (*v1.Condition, bool) {
	if len(conditions) == 0 {
		return &v1.Condition{}, false
	}

	return &conditions[len(conditions)-1], true
}

func (r *SKIPJobReconciler) GetSameConditionIfExists(currentConditions *[]v1.Condition, conditionToFind *v1.Condition) (*v1.Condition, bool) {
	for _, condition := range *currentConditions {
		if condition.Type == conditionToFind.Type {
			return &condition, true
		}
	}

	return nil, false
}

func conditionsHaveSameStatus(condition1 *v1.Condition, condition2 *v1.Condition) bool {
	return condition1.Status == condition2.Status
}

func conditionsHaveSameType(condition1 *v1.Condition, condition2 *v1.Condition) bool {
	return condition1.Type == condition2.Type
}
