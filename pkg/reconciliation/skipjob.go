package reconciliation

import (
	"context"

	skiperatorv1beta1 "github.com/kartverket/skiperator/api/v1beta1"
	"github.com/kartverket/skiperator/internal/config"
	"github.com/kartverket/skiperator/pkg/log"
	"github.com/kartverket/skiperator/pkg/mesh"
	"k8s.io/client-go/rest"
)

type JobReconciliation struct {
	baseReconciliation
}

func NewJobReconciliation(ctx context.Context, job *skiperatorv1beta1.SKIPJob, logger log.Logger, meshMode mesh.Mode, restConfig *rest.Config, skiperatorConfig config.SkiperatorConfig) *JobReconciliation {
	return &JobReconciliation{
		baseReconciliation: baseReconciliation{
			ctx:              ctx,
			logger:           logger,
			meshMode:         meshMode,
			restConfig:       restConfig,
			skipObject:       job,
			skiperatorConfig: skiperatorConfig,
		},
	}
}

func (j *JobReconciliation) GetType() ObjectType {
	return JobType
}
