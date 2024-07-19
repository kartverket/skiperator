package reconciliation

import (
	"context"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/log"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type JobReconciliation struct {
	ctx               context.Context
	job               *skiperatorv1alpha1.SKIPJob
	logger            log.Logger
	resources         []*client.Object
	istioEnabled      bool
	restConfig        *rest.Config
	identityConfigMap *corev1.ConfigMap
}

func NewJobReconciliation(ctx context.Context, job *skiperatorv1alpha1.SKIPJob, logger log.Logger, istioEnabled bool, restConfig *rest.Config, identityConfigMap *corev1.ConfigMap) *JobReconciliation {
	return &JobReconciliation{
		ctx:               ctx,
		job:               job,
		logger:            logger,
		istioEnabled:      istioEnabled,
		restConfig:        restConfig,
		identityConfigMap: identityConfigMap,
	}
}
func (j *JobReconciliation) GetLogger() log.Logger {
	return j.logger
}

func (j *JobReconciliation) GetCtx() context.Context {
	return j.ctx
}

func (j *JobReconciliation) IsIstioEnabled() bool {
	return j.istioEnabled
}

func (j *JobReconciliation) GetReconciliationObject() client.Object {
	return j.job
}

func (j *JobReconciliation) GetType() ReconciliationObjectType {
	return JobType
}

func (j *JobReconciliation) GetRestConfig() *rest.Config {
	return j.restConfig
}

func (j *JobReconciliation) AddResource(object *client.Object) {
	j.resources = append(j.resources, object)
}

func (j *JobReconciliation) GetResources() []*client.Object {
	return j.resources
}

func (j *JobReconciliation) GetCommonSpec() *CommonType {
	return &CommonType{
		GCP:          j.job.Spec.Container.GCP,
		AccessPolicy: j.job.Spec.Container.AccessPolicy,
	}
}

func (j *JobReconciliation) GetIdentityConfigMap() *corev1.ConfigMap {
	return j.identityConfigMap
}
