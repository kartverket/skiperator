package v1alpha1

import (
	"github.com/imdario/mergo"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	DefaultTTLSecondsAfterFinished = int32(60)
)

var JobCreatedCondition = "SKIPJobCreated"

func (skipJob *SKIPJob) ApplyDefaults() error {
	skipJob.setDefaultAnnotations()
	return mergo.Merge(skipJob, getSkipJobDefaults())
}

func getSkipJobDefaults() *SKIPJob {
	return &SKIPJob{
		Spec: SKIPJobSpec{
			Job: &JobSettings{
				// TTLSecondsAfterFinished: &DefaultTTLSecondsAfterFinished,
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
}

func (in *SKIPJob) setDefaultAnnotations() {
	annotations := in.Annotations

	if annotations == nil {
		annotations = map[string]string{}
	}

	// We do not set SyncPolicies if Cron is set. This is due to the recurring nature of Cron jobs not
	// working well in tangent with stuff like deletion policies.
	if in.Spec.Cron == nil && in.Spec.Job != nil {
		if in.Spec.Job.HookSettings != nil {
			// TODO Allow different type of hook delete policies

			annotations["argocd.argoproj.io/hook-delete-policy"] = "HookSucceeded"
			annotations["argocd.argoproj.io/hook"] = *in.Spec.Job.HookSettings.SyncPhase
		}
	}

	in.SetAnnotations(annotations)
}
