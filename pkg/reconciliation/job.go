package reconciliation

import (
	"context"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/log"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type JobReconciliation struct {
	ctx          context.Context
	job          *skiperatorv1alpha1.SKIPJob
	logger       log.Logger
	objects      []client.Object
	istioEnabled bool
	restConfig   *rest.Config
}

func NewJobReconciliation(ctx context.Context, job *skiperatorv1alpha1.SKIPJob, logger log.Logger, restConfig *rest.Config) *JobReconciliation {
	return &JobReconciliation{
		ctx:        ctx,
		job:        job,
		logger:     logger,
		restConfig: restConfig,
	}
}

func (j *JobReconciliation) GetControllerObject() client.Object {
	return j.job
}

func (j *JobReconciliation) GetType() ReconciliationObjectType {
	return JobType
}
