package v1alpha1

import (
	"github.com/imdario/mergo"
	"github.com/kartverket/skiperator/pkg/util"
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
		defaults.Spec.Cron = skipJob.Spec.Cron

		// Due to an error in mergo, bool values are not merged properly. Temporary workaround
		// Should only be necessary for spec.cron.suspend as it's the only bool that we add by default and merge using mergo
		// See: https://github.com/darccio/mergo/issues/237
		//defaults.Spec.Cron.Suspend = util.PointTo(false)
		if *skipJob.Spec.Cron.Suspend {
			defaults.Spec.Cron.Suspend = skipJob.Spec.Cron.Suspend
		} else {
			defaults.Spec.Cron.Suspend = util.PointTo(DefaultSuspend)
		}
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
