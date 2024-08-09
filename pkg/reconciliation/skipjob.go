package reconciliation

import (
	"context"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/log"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
)

type JobReconciliation struct {
	baseReconciliation
}

func NewJobReconciliation(ctx context.Context, job *skiperatorv1alpha1.SKIPJob, logger log.Logger, istioEnabled bool, restConfig *rest.Config, identityConfigMap *corev1.ConfigMap) *JobReconciliation {
	return &JobReconciliation{
		baseReconciliation: baseReconciliation{
			ctx:               ctx,
			logger:            logger,
			istioEnabled:      istioEnabled,
			restConfig:        restConfig,
			identityConfigMap: identityConfigMap,
			skipObject:        job,
		},
	}
}

func (j *JobReconciliation) GetType() ObjectType {
	return JobType
}
