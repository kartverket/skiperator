package job

import (
	"fmt"

	skiperatorv1beta1 "github.com/kartverket/skiperator/api/v1beta1"
	"github.com/kartverket/skiperator/pkg/log"
	"github.com/kartverket/skiperator/pkg/reconciliation"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/gcp"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/pod"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/resourceutils"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/volume"
	"github.com/kartverket/skiperator/pkg/util"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TODO completely butchered, need to be thorougly checked
func Generate(r reconciliation.Reconciliation) error {
	ctxLog := r.GetLogger()
	ctxLog.Debug("Attempting to generate job for skipjob", "skipjob", r.GetSKIPObject().GetName())

	if r.GetType() != reconciliation.JobType {
		return fmt.Errorf("job only supports skipjob type, got %s", r.GetType())
	}

	skipJob := r.GetSKIPObject().(*skiperatorv1beta1.SKIPJob)

	meta := metav1.ObjectMeta{
		Namespace: skipJob.Namespace,
		Name:      skipJob.Name,
		Labels:    make(map[string]string),
	}

	setJobLabels(&ctxLog, skipJob, meta.Labels)
	job := batchv1.Job{ObjectMeta: meta}
	cronJob := batchv1.CronJob{ObjectMeta: meta}

	// By specifying port and path annotations, Istio will scrape metrics from the application
	// and merge it together with its own metrics.
	//
	// See
	//  - https://superorbital.io/blog/istio-metrics-merging/
	//  - https://androidexample365.com/an-example-of-how-istio-metrics-merging-works/
	if r.IsIstioEnabled() && skipJob.Spec.Prometheus != nil {
		skipJob.Annotations["prometheus.io/port"] = skipJob.Spec.Prometheus.Port.StrVal
		skipJob.Annotations["prometheus.io/path"] = skipJob.Spec.Prometheus.Path
	}

	if skipJob.Spec.Cron != nil {
		cronJob.Spec = getCronJobSpec(&ctxLog, skipJob, cronJob.Spec.JobTemplate.Spec.Selector, cronJob.Spec.JobTemplate.Spec.Template.Labels, r.GetIdentityConfigMap())
		r.AddResource(&cronJob)
	} else {
		job.Spec = getJobSpec(&ctxLog, skipJob, job.Spec.Selector, job.Spec.Template.Labels, r.GetIdentityConfigMap())
		r.AddResource(&job)
	}

	return nil
}

func getCronJobSpec(logger *log.Logger, skipJob *skiperatorv1beta1.SKIPJob, selector *metav1.LabelSelector, podLabels map[string]string, gcpIdentityConfigMap *corev1.ConfigMap) batchv1.CronJobSpec {
	spec := batchv1.CronJobSpec{
		Schedule:                skipJob.Spec.Cron.Schedule,
		TimeZone:                skipJob.Spec.Cron.TimeZone,
		StartingDeadlineSeconds: skipJob.Spec.Cron.StartingDeadlineSeconds,
		ConcurrencyPolicy:       skipJob.Spec.Cron.ConcurrencyPolicy,
		Suspend:                 skipJob.Spec.Cron.Suspend,
		JobTemplate: batchv1.JobTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: skipJob.GetDefaultLabels(),
			},
			Spec: getJobSpec(logger, skipJob, selector, podLabels, gcpIdentityConfigMap),
		},
		SuccessfulJobsHistoryLimit: util.PointTo(int32(3)),
		FailedJobsHistoryLimit:     util.PointTo(int32(1)),
	}
	// it's not a default label, maybe it could be?
	// used for selecting workloads by netpols, grafana etc
	setJobLabels(logger, skipJob, spec.JobTemplate.Labels)

	return spec
}

func getJobSpec(logger *log.Logger, skipJob *skiperatorv1beta1.SKIPJob, selector *metav1.LabelSelector, podLabels map[string]string, gcpIdentityConfigMap *corev1.ConfigMap) batchv1.JobSpec {
	podVolumes, containerVolumeMounts := volume.GetContainerVolumeMountsAndPodVolumes(skipJob.Spec.FilesFrom)
	envVars := skipJob.Spec.Env

	if skipJob.Spec.GCP != nil {
		gcpPodVolume := gcp.GetGCPContainerVolume(gcpIdentityConfigMap.Data["workloadIdentityPool"], skipJob.Name)
		gcpContainerVolumeMount := gcp.GetGCPContainerVolumeMount()
		gcpEnvVar := gcp.GetGCPEnvVar()

		podVolumes = append(podVolumes, gcpPodVolume)
		containerVolumeMounts = append(containerVolumeMounts, gcpContainerVolumeMount)
		envVars = append(envVars, gcpEnvVar)
	}

	var skipJobContainer corev1.Container
	skipJobContainer = pod.CreateJobContainer(skipJob, containerVolumeMounts, envVars)

	var containers []corev1.Container

	containers = append(containers, skipJobContainer)

	if util.IsCloudSqlProxyEnabled(skipJob.Spec.GCP) {
		cloudSqlProxyContainer := pod.CreateCloudSqlProxyContainer(skipJob.Spec.GCP.CloudSQLProxy)
		containers = append(containers, cloudSqlProxyContainer)
	}

	jobSpec := batchv1.JobSpec{
		Parallelism:           util.PointTo(int32(1)),
		Completions:           util.PointTo(int32(1)),
		ActiveDeadlineSeconds: skipJob.Spec.Job.ActiveDeadlineSeconds,
		PodFailurePolicy:      nil,
		BackoffLimit:          skipJob.Spec.Job.BackoffLimit,
		Selector:              nil,
		ManualSelector:        nil,
		Template: corev1.PodTemplateSpec{
			Spec: pod.CreatePodSpec(
				containers,
				podVolumes,
				skipJob.KindPostFixedName(),
				skipJob.Spec.Priority,
				skipJob.Spec.RestartPolicy,
				skipJob.Spec.PodSettings,
				skipJob.Name,
			),
			ObjectMeta: metav1.ObjectMeta{
				Labels: skipJob.GetDefaultLabels(),
			},
		},
		TTLSecondsAfterFinished: skipJob.Spec.Job.TTLSecondsAfterFinished,
		CompletionMode:          util.PointTo(batchv1.NonIndexedCompletion),
		Suspend:                 skipJob.Spec.Job.Suspend,
	}

	// it's not a default label, maybe it could be?
	// used for selecting workloads by netpols, grafana etc

	setJobLabels(logger, skipJob, jobSpec.Template.Labels)

	return jobSpec
}

func setJobLabels(logger *log.Logger, skipJob *skiperatorv1beta1.SKIPJob, labels map[string]string) {
	labels["app"] = skipJob.KindPostFixedName()
	labels["app.kubernetes.io/version"] = resourceutils.HumanReadableVersion(logger, skipJob.Spec.Image)
}
