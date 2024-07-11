package job

import (
	"fmt"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/reconciliation"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/gcp"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/pod"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/volume"
	"github.com/kartverket/skiperator/pkg/util"
	"golang.org/x/exp/maps"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

var (
	DefaultAwaitCronJobResourcesWait = time.Second * 10

	SKIPJobReferenceLabelKey = "skiperator.kartverket.no/skipjobName"

	IsSKIPJobKey = "skiperator.kartverket.no/skipjob"
)

// TODO completely butchered, need to be thorougly checked
func Generate(r reconciliation.Reconciliation) error {
	ctxLog := r.GetLogger()
	ctxLog.Debug("Attempting to generate job for skipjob", r.GetReconciliationObject().GetName())

	if r.GetType() != reconciliation.JobType {
		return fmt.Errorf("job only supports skipjob type", r.GetType())
	}

	skipJob := r.GetReconciliationObject().(*skiperatorv1alpha1.SKIPJob)

	job := batchv1.Job{ObjectMeta: metav1.ObjectMeta{
		Namespace: skipJob.Namespace,
		Name:      skipJob.Name,
	}}

	cronJob := batchv1.CronJob{ObjectMeta: metav1.ObjectMeta{
		Namespace: skipJob.Namespace,
		Name:      skipJob.Name,
	}}

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

		ctxLog.Info(fmt.Sprintf("cronjob %v/%v created, requeuing reconcile in %v seconds to await subresource creation", cronJob.Namespace, cronJob.Name, DefaultAwaitCronJobResourcesWait.Seconds()))
		return nil
	}
	if skipJob.Spec.Job != nil {

		desiredSpec := getJobSpec(skipJob, job.Spec.Selector, job.Spec.Template.Labels, r.GetIdentityConfigMap())
		job.Labels = GetJobLabels(skipJob, job.Labels)
		job.Spec = desiredSpec

		return nil

	}

	return nil
}

func getCronJobSpec(skipJob *skiperatorv1alpha1.SKIPJob, selector *metav1.LabelSelector, podLabels map[string]string, gcpIdentityConfigMap *corev1.ConfigMap) batchv1.CronJobSpec {
	return batchv1.CronJobSpec{
		Schedule:                skipJob.Spec.Cron.Schedule,
		StartingDeadlineSeconds: skipJob.Spec.Cron.StartingDeadlineSeconds,
		ConcurrencyPolicy:       skipJob.Spec.Cron.ConcurrencyPolicy,
		Suspend:                 skipJob.Spec.Cron.Suspend,
		JobTemplate: batchv1.JobTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: GetJobLabels(skipJob, podLabels),
			},
			Spec: getJobSpec(skipJob, selector, podLabels, gcpIdentityConfigMap),
		},
		SuccessfulJobsHistoryLimit: util.PointTo(int32(3)),
		FailedJobsHistoryLimit:     util.PointTo(int32(1)),
	}
}

func GetJobLabels(skipJob *skiperatorv1alpha1.SKIPJob, labels map[string]string) map[string]string {
	if len(labels) == 0 {
		labels = make(map[string]string)
	}

	// Used by hahaha to know that the Pod should be watched for killing sidecars
	labels[IsSKIPJobKey] = "true"
	maps.Copy(labels, util.GetPodAppSelector(skipJob.KindPostFixedName()))

	// Added to be able to add the SKIPJob to a reconcile queue when Watched Jobs are queued
	labels[SKIPJobReferenceLabelKey] = skipJob.Name

	return labels
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
				Labels: GetJobLabels(skipJob, nil),
			},
		},
		TTLSecondsAfterFinished: skipJob.Spec.Job.TTLSecondsAfterFinished,
		CompletionMode:          util.PointTo(batchv1.NonIndexedCompletion),
		Suspend:                 skipJob.Spec.Job.Suspend,
	}

	// Jobs create their own selector with a random UUID. Upon creation of the Job we do not know this beforehand.
	// Therefore, simply set these again if they already exist, which would be the case if reconciling an existing job.
	if selector != nil {
		jobSpec.Selector = selector
		if jobSpec.Template.ObjectMeta.Labels == nil {
			jobSpec.Template.ObjectMeta.Labels = map[string]string{}
		}
		maps.Copy(jobSpec.Template.ObjectMeta.Labels, podLabels)
	}

	return jobSpec
}
