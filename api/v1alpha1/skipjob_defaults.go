package v1alpha1

import (
	"dario.cat/mergo"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	DefaultTTLSecondsAfterFinished = int32(60 * 60 * 24 * 7) // One week
	DefaultBackoffLimit            = int32(6)

	DefaultSuspend = false
)

var JobCreatedCondition = "SKIPJobCreated"

func (skipJob *SKIPJob) ApplyDefaults() error {
	skipJob.setDefaultAnnotations()
	return skipJob.setSkipJobDefaults()
}

func (skipJob *SKIPJob) setSkipJobDefaults() error {

	defaults := &SKIPJob{
		Spec: SKIPJobSpec{
			Job: &JobSettings{
				TTLSecondsAfterFinished: &DefaultTTLSecondsAfterFinished,
				BackoffLimit:            &DefaultBackoffLimit,
				Suspend:                 &DefaultSuspend,
			},
		},
		Status: SKIPJobStatus{
			Conditions: []metav1.Condition{
				{
					Type:               JobCreatedCondition,
					Status:             metav1.ConditionTrue,
					LastTransitionTime: metav1.Now(),
					Reason:             "SKIPJobCreated",
					Message:            "SKIPJob was created",
				},
			},
		},
	}

	if skipJob.Spec.Cron != nil {
		defaults.Spec.Cron = &CronSettings{}
		suspend := false
		defaults.Spec.Cron.Suspend = &suspend
	}

	return mergo.Merge(skipJob, defaults)
}

func (skipJob *SKIPJob) setDefaultAnnotations() {
	annotations := skipJob.Annotations

	if annotations == nil {
		annotations = map[string]string{}
	}

	skipJob.SetAnnotations(annotations)
}
