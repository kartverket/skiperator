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
	DefaultPollingRate = time.Second * 30

	DefaultRequeueForPodsWait = time.Second * 15

	DefaultAwaitCronJobResourcesWait = time.Second * 10

	IstioProxyPodContainerName = "istio-proxy"
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
		return reconcile.Result{}, err
	}

	if skipJob.Spec.Cron != nil {
		err = r.GetClient().Get(ctx, types.NamespacedName{
			Namespace: cronJob.Namespace,
			Name:      cronJob.Name,
		}, &cronJob)

		if errors.IsNotFound(err) {
			err := ctrlutil.SetControllerReference(skipJob, &cronJob, r.GetScheme())
			if err != nil {
				return reconcile.Result{}, err
			}

			util.SetCommonAnnotations(&cronJob)

			cronJob.Spec = getCronJobSpec(skipJob, nil, nil, gcpIdentityConfigMap)

			err = r.GetClient().Create(ctx, &cronJob)
			if err != nil {
				return reconcile.Result{}, err
			}

			log.FromContext(ctx).Info(fmt.Sprintf("cronjob %v/%v created, requeuing reconcile in %v seconds to await subresource creation", cronJob.Namespace, cronJob.Name, DefaultAwaitCronJobResourcesWait.Seconds()))
			return reconcile.Result{RequeueAfter: 5}, nil
		} else if err == nil {
			currentSpec := cronJob.Spec
			desiredSpec := getCronJobSpec(skipJob, cronJob.Spec.JobTemplate.Spec.Selector, cronJob.Spec.JobTemplate.Spec.Template.Labels, gcpIdentityConfigMap)

			cronJobSpecDiff, err := util.GetObjectDiff(currentSpec, desiredSpec)
			if err != nil {
				r.EmitWarningEvent(skipJob, "CouldNotUpdateCronJob", fmt.Sprintf("something went wrong when updating the CronJob subresource of SKIPJob %v: %v", skipJob.Name, err))
				return reconcile.Result{}, err
			}

			if len(cronJobSpecDiff) > 0 {
				cronJob.Spec = desiredSpec
				err = r.GetClient().Update(ctx, &cronJob)
				if err != nil {
					r.EmitWarningEvent(skipJob, "CouldNotUpdateCronJob", fmt.Sprintf("something went wrong when updating the CronJob subresource of SKIPJob %v: %v", skipJob.Name, err))
					return reconcile.Result{}, err
				}
			}
		} else if err != nil {
			r.EmitWarningEvent(skipJob, "CouldNotGetCronJob", fmt.Sprintf("something went wrong when getting the CronJob subresource of SKIPJob %v: %v", skipJob.Name, err))
			return reconcile.Result{}, err
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
				return reconcile.Result{}, err
			}

			desiredSpec := getJobSpec(skipJob, job.Spec.Selector, job.Spec.Template.Labels, gcpIdentityConfigMap)
			job.Labels = GetJobLabels(skipJob, job.Labels)
			job.Spec = desiredSpec

			err := r.GetClient().Create(ctx, &job)
			if err != nil {
				r.EmitWarningEvent(skipJob, "CouldNotCreateJob", fmt.Sprintf("something went wrong when creating the Job subresource of SKIPJob %v: %v", skipJob.Name, err))
				return reconcile.Result{}, err
			}

			err = r.SetStatusRunning(ctx, skipJob)

			return reconcile.Result{}, err
		} else if err == nil {
			currentSpec := job.Spec
			desiredSpec := getJobSpec(skipJob, job.Spec.Selector, job.Spec.Template.Labels, gcpIdentityConfigMap)

			jobDiff, err := util.GetObjectDiff(currentSpec, desiredSpec)
			if err != nil {
				r.EmitWarningEvent(skipJob, "CouldNotUpdateJob", fmt.Sprintf("something went wrong when updating the Job subresource of SKIPJob %v: %v", skipJob.Name, err))
				return reconcile.Result{}, err
			}

			if len(jobDiff) > 0 {
				job.Spec = desiredSpec
				err := r.GetClient().Update(ctx, &job)
				if err != nil {
					r.EmitWarningEvent(skipJob, "CouldNotUpdateJob", fmt.Sprintf("something went wrong when updating the Job subresource of SKIPJob %v: %v", skipJob.Name, err))
					return reconcile.Result{}, err
				}

				return reconcile.Result{}, nil
			}
		} else if err != nil {
			r.EmitWarningEvent(skipJob, "CouldNotGetJob", fmt.Sprintf("something went wrong when getting the Job subresource of SKIPJob %v: %v", skipJob.Name, err))
			return reconcile.Result{}, err
		}
	}

	jobsToCheckList := batchv1.JobList{}

	err = r.GetClient().List(ctx, &jobsToCheckList, client.MatchingLabels{
		"app": skipJob.KindPostFixedName(),
	})
	if err != nil {
		return reconcile.Result{}, err
	}

	if len(jobsToCheckList.Items) == 0 {
		log.FromContext(ctx).Info(fmt.Sprintf("could not find any jobs related to SKIPJob %v/%v, skipping job checks", skipJob.Namespace, skipJob.Name))
		return reconcile.Result{}, nil
	}

	for _, job := range jobsToCheckList.Items {
		if isFailed, failedJobMessage := isFailedJob(job); isFailed {
			err = r.SetStatusFailed(ctx, skipJob, fmt.Sprintf("job %v/%v failed, reason:  %v", job.Name, job.Namespace, failedJobMessage))
			continue
		}

		if job.Status.Active > 0 {
			jobPods := corev1.PodList{}
			err := r.GetClient().List(ctx, &jobPods, client.MatchingLabels{"job-name": job.Name})
			if err != nil {
				r.EmitWarningEvent(skipJob, "CouldNotListPods", fmt.Sprintf("something went wrong when listing Pods of Job %v: %v", job.Name, err))
				return reconcile.Result{}, err
			}

			if len(jobPods.Items) == 0 {
				// In the case that the job has no pods yet, we should requeue the request so that the controller can
				// check the pods when created.
				log.FromContext(ctx).Info(fmt.Sprintf("could not find pods for job %v/%v, requeuing reconcile in %v seconds", job.Namespace, job.Name, DefaultRequeueForPodsWait.Seconds()))
				return reconcile.Result{RequeueAfter: DefaultRequeueForPodsWait}, nil
			}

			for _, pod := range jobPods.Items {
				if pod.Status.Phase == corev1.PodFailed {
					err := r.SetStatusFailed(ctx, skipJob, fmt.Sprintf("workload container for pod %v failed", pod.Name))
					return reconcile.Result{}, err
				}

				if pod.Status.Phase == corev1.PodPending {
					infoMessage := fmt.Sprintf("pod %v for job %v/%v pending, requeueing in %v seconds", pod.Name, job.Namespace, job.Name, DefaultRequeueForPodsWait.Seconds())
					log.FromContext(ctx).Info(infoMessage)
					return reconcile.Result{RequeueAfter: DefaultRequeueForPodsWait}, nil
				}

				terminatedContainerStatuses := map[string]int32{}
				for _, containerStatus := range pod.Status.ContainerStatuses {
					if containerStatus.State.Terminated != nil {
						terminatedContainerStatuses[containerStatus.Name] = containerStatus.State.Terminated.ExitCode
					}
				}

				if _, exists := terminatedContainerStatuses[IstioProxyPodContainerName]; exists {
					// We want to skip all further operations if the istio-proxy is terminated
					continue
				}

				if exitCode, exists := terminatedContainerStatuses[skipJob.KindPostFixedName()]; exists {
					if exitCode == 0 {
						// Once we know the istio-proxy pod is marked as completed, we can assume the Job is finished
						err = r.SetStatusFinished(ctx, skipJob)
						if err != nil {
							return reconcile.Result{}, err
						}
					} else {
						// TODO Should we set Failed status here, or rather do it based on the job condition?
						err := r.SetStatusFailed(ctx, skipJob, fmt.Sprintf("workload container for pod %v failed with exit code %v", pod.Name, exitCode))
						if err != nil {
							return reconcile.Result{}, err
						}

						infoMessage := fmt.Sprintf("pod %v for job %v/%v failed", pod.Name, job.Namespace, job.Name)
						log.FromContext(ctx).Info(infoMessage)
					}
				} else {
					err := r.SetStatusRunning(ctx, skipJob)
					if err != nil {
						return reconcile.Result{}, err
					}

					infoMessage := fmt.Sprintf("job %v/%v not complete, requeueing in %v seconds", job.Namespace, job.Name, DefaultPollingRate.Seconds())
					log.FromContext(ctx).Info(infoMessage)

					return reconcile.Result{RequeueAfter: DefaultPollingRate}, nil
				}
			}
		}
	}

	return reconcile.Result{}, nil
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
	labels["skiperator.kartverket.no/skipjob"] = "true"
	maps.Copy(labels, util.GetPodAppSelector(skipJob.KindPostFixedName()))

	return labels
}

func getJobSpec(skipJob *skiperatorv1alpha1.SKIPJob, selector *metav1.LabelSelector, podLabels map[string]string, gcpIdentityConfigMap *corev1.ConfigMap) batchv1.JobSpec {
	podVolumes, containerVolumeMounts := core.GetContainerVolumeMountsAndPodVolumes(skipJob.Spec.Container.FilesFrom)

	if skipJob.Spec.Container.GCP != nil {
		gcpPodVolume := gcp.GetGCPContainerVolume(gcpIdentityConfigMap.Data["workloadIdentityPool"], skipJob.Name)
		gcpContainerVolumeMount := gcp.GetGCPContainerVolumeMount()
		gcpEnvVar := gcp.GetGCPEnvVar()

		podVolumes = append(podVolumes, gcpPodVolume)
		containerVolumeMounts = append(containerVolumeMounts, gcpContainerVolumeMount)
		skipJob.Spec.Container.Env = append(skipJob.Spec.Container.Env, gcpEnvVar)
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
			Spec: core.CreatePodSpec(core.CreateJobContainer(skipJob, containerVolumeMounts), podVolumes, skipJob.KindPostFixedName(), skipJob.Spec.Container.Priority, skipJob.Spec.Container.RestartPolicy),
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
