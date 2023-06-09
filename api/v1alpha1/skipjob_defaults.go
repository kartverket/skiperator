package v1alpha1

import (
	"github.com/imdario/mergo"
)

const ()

func (skipJob *SKIPJob) ApplyDefaults() error {
	return mergo.Merge(skipJob, getSkipJobDefaults())
}

func getSkipJobDefaults() *SKIPJob {
	return &SKIPJob{
		Spec: SKIPJobSpec{
			Job: &JobSettings{
				ActiveDeadlineSeconds:   nil,
				BackoffLimit:            nil,
				Parallelism:             nil,
				Suspend:                 nil,
				TTLSecondsAfterFinished: nil,
			},
			Cron: nil,
		},
	}
}
