package v1alpha1

import (
	"github.com/imdario/mergo"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	DefaultTTLSecondsAfterFinished = int32(60)
)

func (skipJob *SKIPJob) ApplyDefaults() error {
	skipJob.SetAnnotations(skipJob.getDefaultAnnotations())
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
			Conditions: []metav1.Condition{},
		},
	}
}

func (in *SKIPJob) getDefaultAnnotations() map[string]string {
	annotations := in.Annotations

	// We do not set SyncPolicies if Cron is set. This is due to the recurring nature of Cron jobs not
	// working well in tangent with stuff like deletion policies.
	if in.Spec.Cron == nil && in.Spec.Job.HookSettings != nil {
		// TODO Allow different type of hook delete policies
		println("hello???")
		annotations["argocd.argoproj.io/hook-delete-policy"] = "HookSucceeded"
		annotations["argocd.argoproj.io/hook"] = *in.Spec.Job.HookSettings.SyncPhase
	}

	return annotations
}
