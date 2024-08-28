package job

import (
	"fmt"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/reconciliation"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/gcp"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/pod"
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

	skipJob := r.GetSKIPObject().(*skiperatorv1alpha1.SKIPJob)

	meta := metav1.ObjectMeta{
		Namespace: skipJob.Namespace,
		Name:      skipJob.Name,
		Labels:    map[string]string{"app": skipJob.KindPostFixedName()},
	}
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
		cronJob.Spec = getCronJobSpec(skipJob, cronJob.Spec.JobTemplate.Spec.Selector, cronJob.Spec.JobTemplate.Spec.Template.Labels, r.GetIdentityConfigMap())
		r.AddResource(&cronJob)
	} else {
		job.Spec = getJobSpec(skipJob, job.Spec.Selector, job.Spec.Template.Labels, r.GetIdentityConfigMap())
		r.AddResource(&job)
	}

	return nil
}

func getCronJobSpec(skipJob *skiperatorv1alpha1.SKIPJob, selector *metav1.LabelSelector, podLabels map[string]string, gcpIdentityConfigMap *corev1.ConfigMap) batchv1.CronJobSpec {
	spec := batchv1.CronJobSpec{
		Schedule:                skipJob.Spec.Cron.Schedule,
		StartingDeadlineSeconds: skipJob.Spec.Cron.StartingDeadlineSeconds,
		ConcurrencyPolicy:       skipJob.Spec.Cron.ConcurrencyPolicy,
		Suspend:                 skipJob.Spec.Cron.Suspend,
		JobTemplate: batchv1.JobTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: skipJob.GetDefaultLabels(),
			},
			Spec: getJobSpec(skipJob, selector, podLabels, gcpIdentityConfigMap),
		},
		SuccessfulJobsHistoryLimit: util.PointTo(int32(3)),
		FailedJobsHistoryLimit:     util.PointTo(int32(1)),
	}
	// it's not a default label, maybe it could be?
	// used for selecting workloads by netpols, grafana etc
	spec.JobTemplate.Labels["app"] = skipJob.KindPostFixedName()

	return spec
}

func getJobSpec(skipJob *skiperatorv1alpha1.SKIPJob, selector *metav1.LabelSelector, podLabels map[string]string, gcpIdentityConfigMap *corev1.ConfigMap) batchv1.JobSpec {
	podVolumes, containerVolumeMounts := volume.GetContainerVolumeMountsAndPodVolumes(skipJob.Spec.Container.FilesFrom)
	envVars := skipJob.Spec.Container.Env

	if skipJob.Spec.Container.GCP != nil {
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
				skipJob.Spec.Container.Priority,
				skipJob.Spec.Container.RestartPolicy,
				skipJob.Spec.Container.PodSettings,
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
	jobSpec.Template.Labels["app"] = skipJob.KindPostFixedName()

	return jobSpec
}
