package v1alpha1

import (
	"github.com/imdario/mergo"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	DefaultTTLSecondsAfterFinished = int32(60)
)

func (skipJob *SKIPJob) ApplyDefaults() error {
	return mergo.Merge(skipJob, getSkipJobDefaults())
}

func getSkipJobDefaults() *SKIPJob {
	return &SKIPJob{
		Status: SKIPJobStatus{
			Conditions: []metav1.Condition{},
		},
		Spec: SKIPJobSpec{
			Job: &JobSettings{
				// TTLSecondsAfterFinished: &DefaultTTLSecondsAfterFinished,
			},
		},
	}
}
