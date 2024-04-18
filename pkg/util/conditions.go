package util

import (
	"context"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"time"
)

type ConditionsAware interface {
	GetConditions() []metav1.Condition
	SetConditions(conditions []metav1.Condition)
}

func AppendCondition(ctx context.Context, reconcilerClient client.Client, object client.Object,
	typeName string, status metav1.ConditionStatus, reason string, message string) error {
	logger := log.FromContext(ctx)

	conditionsAware, conversionSuccessful := (object).(ConditionsAware)
	if conversionSuccessful {
		timeNow := metav1.Time{Time: time.Now()}
		condition := metav1.Condition{Type: typeName, Status: status, Reason: reason, Message: message, LastTransitionTime: timeNow}
		conditionsAware.SetConditions(append(conditionsAware.GetConditions(), condition))
		err := reconcilerClient.Status().Update(ctx, object)
		if err != nil {
			errMessage := "Custom resource status update failed"
			logger.Info(errMessage)
			return fmt.Errorf(errMessage)
		}

	} else {
		errMessage := "Status cannot be set, resource doesn't support conditions"
		logger.Info(errMessage)
		return fmt.Errorf(errMessage)
	}
	return nil
}
