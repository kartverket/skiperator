package skipjobcontroller

import (
	"context"
	"fmt"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/core"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/gcp"
	"github.com/kartverket/skiperator/pkg/util"
	"golang.org/x/exp/maps"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"time"
)

var (
	DefaultAwaitCronJobResourcesWait = time.Second * 10

	SKIPJobReferenceLabelKey = "skiperator.kartverket.no/skipjobName"

	IsSKIPJobKey = "skiperator.kartverket.no/skipjob"
)

func (r *SKIPJobReconciler) reconcileJob(ctx context.Context, skipJob *skiperatorv1alpha1.SKIPJob) (reconcile.Result, error) {
	job := batchv1.Job{ObjectMeta: metav1.ObjectMeta{
		Namespace: skipJob.Namespace,
		Name:      skipJob.Name,
	}}

	cronJob := batchv1.CronJob{ObjectMeta: metav1.ObjectMeta{
		Namespace: skipJob.Namespace,
		Name:      skipJob.Name,
	}}

	gcpIdentityConfigMap, err := r.getGCPIdentityConfigMap(ctx, *skipJob)
	if err != nil {
		return util.RequeueWithError(err)
	}
	// By specifying port and path annotations, Istio will scrape metrics from the application
	// and merge it together with its own metrics.
	//
	// See
	//  - https://superorbital.io/blog/istio-metrics-merging/
	//  - https://androidexample365.com/an-example-of-how-istio-metrics-merging-works/
	istioEnabled := r.IsIstioEnabledForNamespace(ctx, skipJob.Namespace)
	if istioEnabled && skipJob.Spec.Prometheus != nil {
		skipJob.Annotations["prometheus.io/port"] = skipJob.Spec.Prometheus.Port.StrVal
		skipJob.Annotations["prometheus.io/path"] = skipJob.Spec.Prometheus.Path
	}

	if skipJob.Spec.Cron != nil {
		err = r.GetClient().Get(ctx, types.NamespacedName{
			Namespace: cronJob.Namespace,
			Name:      cronJob.Name,
		}, &cronJob)

		if errors.IsNotFound(err) {
			err := ctrlutil.SetControllerReference(skipJob, &cronJob, r.GetScheme())
			if err != nil {
				return util.RequeueWithError(err)
			}

			util.SetCommonAnnotations(&cronJob)

			cronJob.Spec = getCronJobSpec(skipJob, nil, nil, gcpIdentityConfigMap)

			err = r.GetClient().Create(ctx, &cronJob)
			if err != nil {
				return util.RequeueWithError(err)
			}

			log.FromContext(ctx).Info(fmt.Sprintf("cronjob %v/%v created, requeuing reconcile in %v seconds to await subresource creation", cronJob.Namespace, cronJob.Name, DefaultAwaitCronJobResourcesWait.Seconds()))
			return reconcile.Result{RequeueAfter: 5}, nil
		} else if err == nil {
			currentSpec := cronJob.Spec
			desiredSpec := getCronJobSpec(skipJob, cronJob.Spec.JobTemplate.Spec.Selector, cronJob.Spec.JobTemplate.Spec.Template.Labels, gcpIdentityConfigMap)

			cronJobSpecDiff, err := util.GetObjectDiff(currentSpec, desiredSpec)
			if err != nil {
				r.EmitWarningEvent(skipJob, "CouldNotUpdateCronJob", fmt.Sprintf("something went wrong when updating the CronJob subresource of SKIPJob %v: %v", skipJob.Name, err))
				return util.RequeueWithError(err)
			}

			if len(cronJobSpecDiff) > 0 {
				cronJob.Spec = desiredSpec
				err = r.GetClient().Update(ctx, &cronJob)
				if err != nil {
					r.EmitWarningEvent(skipJob, "CouldNotUpdateCronJob", fmt.Sprintf("something went wrong when updating the CronJob subresource of SKIPJob %v: %v", skipJob.Name, err))
					return util.RequeueWithError(err)
				}
			}
		} else if err != nil {
			r.EmitWarningEvent(skipJob, "CouldNotGetCronJob", fmt.Sprintf("something went wrong when getting the CronJob subresource of SKIPJob %v: %v", skipJob.Name, err))
			return util.RequeueWithError(err)
		}
	} else {
		err = r.GetClient().Get(ctx, types.NamespacedName{
			Namespace: skipJob.Namespace,
			Name:      job.Name,
		}, &job)

		if errors.IsNotFound(err) {
			util.SetCommonAnnotations(&job)

			err = ctrlutil.SetControllerReference(skipJob, &job, r.GetScheme())
			if err != nil {
				return util.RequeueWithError(err)
			}

			desiredSpec := getJobSpec(skipJob, job.Spec.Selector, job.Spec.Template.Labels, gcpIdentityConfigMap)
			job.Labels = GetJobLabels(skipJob, job.Labels)
			job.Spec = desiredSpec

			err := r.GetClient().Create(ctx, &job)
			if err != nil {
				r.EmitWarningEvent(skipJob, "CouldNotCreateJob", fmt.Sprintf("something went wrong when creating the Job subresource of SKIPJob %v: %v", skipJob.Name, err))
				return util.RequeueWithError(err)
			}

			err = r.SetStatusRunning(ctx, skipJob)

			return util.RequeueWithError(err)
		} else if err == nil {
			currentSpec := job.Spec
			desiredSpec := getJobSpec(skipJob, job.Spec.Selector, job.Spec.Template.Labels, gcpIdentityConfigMap)

			jobDiff, err := util.GetObjectDiff(currentSpec, desiredSpec)
			if err != nil {
				r.EmitWarningEvent(skipJob, "CouldNotUpdateJob", fmt.Sprintf("something went wrong when updating the Job subresource of SKIPJob %v: %v", skipJob.Name, err))
				return util.RequeueWithError(err)
			}

			if len(jobDiff) > 0 {
				job.Spec = desiredSpec
				err := r.GetClient().Update(ctx, &job)
				if err != nil {
					r.EmitWarningEvent(skipJob, "CouldNotUpdateJob", fmt.Sprintf("something went wrong when updating the Job subresource of SKIPJob %v: %v", skipJob.Name, err))
					return util.RequeueWithError(err)
				}

				return util.DoNotRequeue()
			}
		} else if err != nil {
			r.EmitWarningEvent(skipJob, "CouldNotGetJob", fmt.Sprintf("something went wrong when getting the Job subresource of SKIPJob %v: %v", skipJob.Name, err))
			return util.RequeueWithError(err)
		}
	}

	jobsToCheckList := batchv1.JobList{}

	err = r.GetClient().List(ctx, &jobsToCheckList, client.MatchingLabels{
		SKIPJobReferenceLabelKey: skipJob.Name,
	})
	if err != nil {
		return util.RequeueWithError(err)
	}

	if len(jobsToCheckList.Items) == 0 {
		log.FromContext(ctx).Info(fmt.Sprintf("could not find any jobs related to SKIPJob %v/%v, skipping job checks", skipJob.Namespace, skipJob.Name))
		return util.DoNotRequeue()
	}

	for _, job := range jobsToCheckList.Items {
		if isFailed, failedJobMessage := isFailedJob(job); isFailed {
			err = r.SetStatusFailed(ctx, skipJob, fmt.Sprintf("job %v/%v failed, reason:  %v", job.Name, job.Namespace, failedJobMessage))
			if err != nil {
				return util.RequeueWithError(err)
			}
			continue
		}

		if job.Status.CompletionTime != nil {
			err = r.SetStatusFinished(ctx, skipJob)
			if err != nil {
				return util.RequeueWithError(err)
			}
			continue
		}

		err := r.SetStatusRunning(ctx, skipJob)
		if err != nil {
			return util.RequeueWithError(err)
		}
	}

	return util.DoNotRequeue()
}

func isFailedJob(job batchv1.Job) (bool, string) {
	for _, condition := range job.Status.Conditions {
		if condition.Type == ConditionFailed && condition.Status == corev1.ConditionTrue {
			return true, condition.Message
		}
	}

	return false, ""
}

func (r *SKIPJobReconciler) getGCPIdentityConfigMap(ctx context.Context, skipJob skiperatorv1alpha1.SKIPJob) (*corev1.ConfigMap, error) {
	if skipJob.Spec.Container.GCP != nil {
		gcpIdentityConfigMapNamespacedName := types.NamespacedName{Namespace: "skiperator-system", Name: "gcp-identity-config"}

		configMap, err := util.GetConfigMap(r.GetClient(), ctx, gcpIdentityConfigMapNamespacedName)
		if !util.ErrIsMissingOrNil(
			r.GetRecorder(),
			err,
			"Cannot find configmap named "+gcpIdentityConfigMapNamespacedName.Name+" in namespace "+gcpIdentityConfigMapNamespacedName.Namespace,
			&skipJob,
		) {
			return nil, err
		}

		return &configMap, nil
	} else {
		return nil, nil
	}
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
	podVolumes, containerVolumeMounts := core.GetContainerVolumeMountsAndPodVolumes(skipJob.Spec.Container.FilesFrom)
	envVars := skipJob.Spec.Container.Env

	if skipJob.Spec.Container.GCP != nil {
		gcpPodVolume := gcp.GetGCPContainerVolume(gcpIdentityConfigMap.Data["workloadIdentityPool"], skipJob.Name)
		gcpContainerVolumeMount := gcp.GetGCPContainerVolumeMount()
		gcpEnvVar := gcp.GetGCPEnvVar()

		podVolumes = append(podVolumes, gcpPodVolume)
		containerVolumeMounts = append(containerVolumeMounts, gcpContainerVolumeMount)
		envVars = append(envVars, gcpEnvVar)
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
			Spec: core.CreatePodSpec(
				core.CreateJobContainer(skipJob, containerVolumeMounts, envVars),
				podVolumes,
				skipJob.KindPostFixedName(),
				skipJob.Spec.Container.Priority,
				skipJob.Spec.Container.RestartPolicy,
				skipJob.Spec.Container.PodSettings,
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
