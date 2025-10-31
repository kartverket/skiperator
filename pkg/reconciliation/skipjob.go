package reconciliation

import (
	"context"

	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/internal/config"
	"github.com/kartverket/skiperator/pkg/log"
	"k8s.io/client-go/rest"
)

type JobReconciliation struct {
	baseReconciliation
}

func NewJobReconciliation(ctx context.Context, job *skiperatorv1alpha1.SKIPJob, logger log.Logger, istioEnabled bool, restConfig *rest.Config, skiperatorConfig config.SkiperatorConfig) *JobReconciliation {
	return &JobReconciliation{
		baseReconciliation: baseReconciliation{
			ctx:              ctx,
			logger:           logger,
			istioEnabled:     istioEnabled,
			restConfig:       restConfig,
			skipObject:       job,
			skiperatorConfig: skiperatorConfig,
		},
	}
}

func (j *JobReconciliation) GetType() ObjectType {
	return JobType
}
