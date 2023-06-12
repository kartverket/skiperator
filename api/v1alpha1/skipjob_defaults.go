package v1alpha1

import (
	"github.com/imdario/mergo"
	"github.com/kartverket/skiperator/pkg/util"
)

const (
	DefaultTTLSecondsAfterFinished = int32(60)
)

func (skipJob *SKIPJob) ApplyDefaults() error {
	return mergo.Merge(skipJob, getSkipJobDefaults())
}

func getSkipJobDefaults() *SKIPJob {
	return &SKIPJob{
		Spec: SKIPJobSpec{
			Job: &JobSettings{
				TTLSecondsAfterFinished: util.PointTo(DefaultTTLSecondsAfterFinished),
			},
			Cron: &CronSettings{
				ConcurrencyPolicy:       "",
				Schedule:                "",
				StartingDeadlineSeconds: nil,
				Suspend:                 nil,
			},
		},
	}
}
