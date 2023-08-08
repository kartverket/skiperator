package skipjobcontroller

import (
	"context"
	"encoding/json"
	"fmt"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/core"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/gcp"
	"github.com/kartverket/skiperator/pkg/util"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"time"
)

var (
	SKIPJobReferenceLabelKey = "skipJobOwnerName"
	DefaultPollingRate       = time.Second * 15

	IstioProxyPodContainerName = "istio-proxy"
)

func (r *SKIPJobReconciler) reconcileJob(ctx context.Context, skipJob *skiperatorv1alpha1.SKIPJob) (reconcile.Result, error) {
	job := batchv1.Job{ObjectMeta: metav1.ObjectMeta{
		Namespace: skipJob.Namespace,
		Name:      skipJob.HashedName(),
	}}

	cronJob := batchv1.CronJob{ObjectMeta: metav1.ObjectMeta{
		Namespace: skipJob.Namespace,
		Name:      skipJob.HashedName(),
	}}

	gcpIdentityConfigMap, err := r.getGCPIdentityConfigMap(ctx, *skipJob)
	if err != nil {
		return reconcile.Result{}, err
	}

	if skipJob.Spec.Cron != nil {
		err := deleteJobIfExists(r.GetClient(), ctx, job)
		if err != nil {
			r.EmitWarningEvent(skipJob, "CouldNotDeleteJob", fmt.Sprintf("something went wrong when deleting the outdated Job subresource of SKIPJob %v: %v", skipJob.Name, err))
			return reconcile.Result{}, err
		}

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

			cronJob.Spec = getCronJobSpec(skipJob, cronJob.Name, cronJob.Spec.JobTemplate.Spec.Selector, cronJob.Labels, gcpIdentityConfigMap)

			err = r.GetClient().Create(ctx, &cronJob)
			if err != nil {
				return reconcile.Result{}, err
			}

			// Requeue the reconciliation after creating the Cron Job to allow the control plane to create subresources
			return reconcile.Result{
				Requeue:      true,
				RequeueAfter: 5,
			}, nil
		} else if err == nil {
			currentSpec := cronJob.Spec
			desiredSpec := getCronJobSpec(skipJob, cronJob.Name, cronJob.Spec.JobTemplate.Spec.Selector, cronJob.Labels, gcpIdentityConfigMap)

			jobTemplateDiff, err := util.GetObjectDiff(currentSpec.JobTemplate, desiredSpec.JobTemplate)
			if err != nil {
				r.EmitWarningEvent(skipJob, "CouldNotUpdateCronJob", fmt.Sprintf("something went wrong when updating the CronJob subresource of SKIPJob %v: %v", skipJob.Name, err))
				return reconcile.Result{}, err
			}

			if len(jobTemplateDiff) > 0 {
				r.EmitWarningEvent(skipJob, "CantUpdateCronJob", fmt.Sprintf("an attempt was made to update the job template of CronJob %v, CronJob must be deleted to update the job template", cronJob.Name))
				return reconcile.Result{}, err
			}

			fullJobDiff, err := util.GetObjectDiff(currentSpec, desiredSpec)
			if err != nil {
				r.EmitWarningEvent(skipJob, "CouldNotUpdateCronJob", fmt.Sprintf("something went wrong when updating the CronJob subresource of SKIPJob %v: %v", skipJob.Name, err))
				return reconcile.Result{}, err
			}

			if len(fullJobDiff) > 0 {
				patch := client.MergeFrom(cronJob.DeepCopy())
				err = r.GetClient().Patch(ctx, &cronJob, patch)
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
		err := deleteCronJobIfExists(r.GetClient(), ctx, cronJob)
		if err != nil {
			r.EmitWarningEvent(skipJob, "CouldNotDeleteCronJob", fmt.Sprintf("something went wrong when deleting the outdated CronJob subresource of SKIPJob %v: %v", skipJob.Name, err))
			return reconcile.Result{}, err
		}

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

			wantedSpec := getJobSpec(skipJob, job.Spec.Selector, job.Labels, gcpIdentityConfigMap)
			job.Labels = GetJobLabels(skipJob, job.Name, job.Labels)
			job.Spec = wantedSpec

			err := r.GetClient().Create(ctx, &job)
			if err != nil {
				r.EmitWarningEvent(skipJob, "CouldNotCreateJob", fmt.Sprintf("something went wrong when creating the Job subresource of SKIPJob %v: %v", skipJob.Name, err))
				return reconcile.Result{}, err
			}

			err = r.SetStatusRunning(ctx, skipJob)

			return reconcile.Result{}, err
		} else if err == nil {
			currentSpec := job.Spec
			desiredSpec := getJobSpec(skipJob, job.Spec.Selector, job.Labels, gcpIdentityConfigMap)

			jobDiff, err := util.GetObjectDiff(currentSpec, desiredSpec)
			if err != nil {
				r.EmitWarningEvent(skipJob, "CouldNotUpdateJob", fmt.Sprintf("something went wrong when updating the Job subresource of SKIPJob %v: %v", skipJob.Name, err))
				return reconcile.Result{}, err
			}

			if len(jobDiff) > 0 {
				r.EmitWarningEvent(skipJob, "CantUpdateJob", fmt.Sprintf("an attempt was made to update the Job subresource of SKIPJob %v, Jobs cannot be updated after creation and must be deleted to update", skipJob.Name))
				return reconcile.Result{}, err
			}
		} else if err != nil {
			r.EmitWarningEvent(skipJob, "CouldNotGetJob", fmt.Sprintf("something went wrong when getting the Job subresource of SKIPJob %v: %v", skipJob.Name, err))
			return reconcile.Result{}, err
		}
	}

	jobsToCheckList := batchv1.JobList{}

	err = r.GetClient().List(ctx, &jobsToCheckList, client.MatchingLabels{
		SKIPJobReferenceLabelKey: skipJob.Name,
	})
	if err != nil {
		return reconcile.Result{}, err
	}

	for _, job := range jobsToCheckList.Items {
		if job.Status.CompletionTime == nil {
			jobPods := corev1.PodList{}
			err := r.GetClient().List(ctx, &jobPods, client.MatchingLabels{"job-name": job.Name})
			if err != nil {
				r.EmitWarningEvent(skipJob, "CouldNotListPods", fmt.Sprintf("something went wrong when listing Pods of Job %v: %v", job.Name, err))
				return reconcile.Result{}, err
			}

			if len(jobPods.Items) == 0 {
				// In the case that the job has no pods yet, we should requeue the request so that the controller can
				// check the pods when created.
				log.FromContext(ctx).Info(fmt.Sprintf("could not find pods for job %v/%v, requeuing reconcile", job.Name, job.Namespace))
				return reconcile.Result{RequeueAfter: time.Second * 5}, nil
			}

			for _, pod := range jobPods.Items {
				if pod.Status.Phase == corev1.PodFailed {
					err := r.SetStatusFailed(ctx, skipJob, fmt.Sprintf("workload container for pod %v failed", pod.Name))
					return reconcile.Result{}, err
				}

				if pod.Status.Phase != corev1.PodRunning {
					continue
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

				if exitCode, exists := terminatedContainerStatuses[job.Labels["job-name"]]; exists {
					if exitCode == 0 {
						ephemeralContainerPatch, err := getEphemeralContainerPatch(pod)
						if err != nil {
							r.EmitWarningEvent(skipJob, "CouldNotCreateEphemeralContainer", fmt.Sprintf("something went wrong when creating ephemeral istio killer-container for Job %v: %v", job.Name, err))
							return reconcile.Result{}, err
						}

						err = r.GetClient().SubResource("ephemeralcontainers").Patch(ctx, &pod, client.RawPatch(types.StrategicMergePatchType, ephemeralContainerPatch))
						if err != nil {
							r.EmitWarningEvent(skipJob, "CouldNotExitContainer", fmt.Sprintf("something went wrong when killing istio container for Job %v: %v", job.Name, err))
							return reconcile.Result{}, err
						}

						// Once we know the istio-proxy pod is marked as completed, we can assume the Job is finished
						err = r.SetStatusFinished(ctx, skipJob)
						if err != nil {
							return reconcile.Result{}, err
						}
					} else {
						err := r.SetStatusFailed(ctx, skipJob, fmt.Sprintf("workload container for pod %v failed with exit code %v", pod.Name, exitCode))
						return reconcile.Result{}, err
					}
				} else {
					err := r.SetStatusRunning(ctx, skipJob)
					if err != nil {
						return reconcile.Result{}, err
					}

					infoMessage := fmt.Sprintf("job %v/%v not complete, requeueing in %v seconds", job.Name, job.Namespace, DefaultPollingRate.Seconds())
					log.FromContext(ctx).Info(infoMessage)

					return reconcile.Result{RequeueAfter: DefaultPollingRate}, nil
				}
			}
		} else {
			// RFC: Do we need to do anything when jobs are finished?
			// Perhaps remove them after x time? CronJobs automatically clear old jobs for one
		}
	}

	return reconcile.Result{}, nil
}

func getEphemeralContainerPatch(pod corev1.Pod) ([]byte, error) {
	tempPod, _, err := generateDebugContainer(&pod)
	if err != nil {
		return nil, err
	}

	podJSON, _ := json.Marshal(pod)
	ecPodJSON, _ := json.Marshal(tempPod)

	patch, err := strategicpatch.CreateTwoWayMergePatch(podJSON, ecPodJSON, pod)

	return patch, err
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

func deleteCronJobIfExists(recClient client.Client, context context.Context, cronJob batchv1.CronJob) error {
	err := recClient.Delete(context, &cronJob)
	err = client.IgnoreNotFound(err)
	if err != nil {

		return err
	}

	return nil
}

func deleteJobIfExists(recClient client.Client, context context.Context, job batchv1.Job) error {
	err := recClient.Delete(context, &job)
	err = client.IgnoreNotFound(err)
	if err != nil {

		return err
	}

	return nil
}

func getCronJobSpec(skipJob *skiperatorv1alpha1.SKIPJob, jobName string, selector *metav1.LabelSelector, labels map[string]string, gcpIdentityConfigMap *corev1.ConfigMap) batchv1.CronJobSpec {
	return batchv1.CronJobSpec{
		Schedule:                skipJob.Spec.Cron.Schedule,
		StartingDeadlineSeconds: skipJob.Spec.Cron.StartingDeadlineSeconds,
		ConcurrencyPolicy:       skipJob.Spec.Cron.ConcurrencyPolicy,
		Suspend:                 skipJob.Spec.Cron.Suspend,
		JobTemplate: batchv1.JobTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: GetJobLabels(skipJob, jobName, labels),
			},
			Spec: getJobSpec(skipJob, selector, labels, gcpIdentityConfigMap),
		},
		SuccessfulJobsHistoryLimit: util.PointTo(int32(3)),
		FailedJobsHistoryLimit:     util.PointTo(int32(1)),
	}
}

func GetJobLabels(skipJob *skiperatorv1alpha1.SKIPJob, jobName string, labels map[string]string) map[string]string {
	if len(labels) == 0 {
		labels = make(map[string]string)
	}
	labels["job-name"] = jobName
	labels[SKIPJobReferenceLabelKey] = skipJob.Name

	return labels
}

func getJobSpec(skipJob *skiperatorv1alpha1.SKIPJob, selector *metav1.LabelSelector, labels map[string]string, gcpIdentityConfigMap *corev1.ConfigMap) batchv1.JobSpec {
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
			Spec: core.CreatePodSpec(core.CreateJobContainer(skipJob, containerVolumeMounts), podVolumes, skipJob.HashedName(), skipJob.Spec.Container.Priority, skipJob.Spec.Container.RestartPolicy),
		},
		TTLSecondsAfterFinished: skipJob.Spec.Job.TTLSecondsAfterFinished,
		CompletionMode:          util.PointTo(batchv1.NonIndexedCompletion),
		Suspend:                 skipJob.Spec.Job.Suspend,
	}

	// Jobs create their own selector with a random UUID. Upon creation of the Job we do not know this beforehand.
	// Therefore, simply set these again if they already exist, which would be the case if reconciling an existing job.
	if selector != nil {
		jobSpec.Selector = selector
		jobSpec.Template.ObjectMeta.Labels = labels
	}

	return jobSpec
}

func generateDebugContainer(pod *corev1.Pod) (*corev1.Pod, *corev1.EphemeralContainer, error) {
	ec := &corev1.EphemeralContainer{
		EphemeralContainerCommon: corev1.EphemeralContainerCommon{
			Name:                     "istio-debugger",
			Image:                    "istio/base",
			Command:                  []string{"/bin/sh", "-c", "curl --max-time 2 -s -f -XPOST http://127.0.0.1:15000/quitquitquit"},
			TerminationMessagePolicy: corev1.TerminationMessageReadFile,
			SecurityContext: util.PointTo(corev1.SecurityContext{
				RunAsUser:                util.PointTo(int64(1337)),
				RunAsGroup:               util.PointTo(int64(1337)),
				RunAsNonRoot:             util.PointTo(true),
				ReadOnlyRootFilesystem:   util.PointTo(true),
				AllowPrivilegeEscalation: util.PointTo(false),
				SeccompProfile: util.PointTo(corev1.SeccompProfile{
					Type: corev1.SeccompProfileTypeRuntimeDefault,
				}),
			}),
		},
		TargetContainerName: IstioProxyPodContainerName,
	}

	copied := pod.DeepCopy()
	copied.Spec.EphemeralContainers = append(copied.Spec.EphemeralContainers, *ec)

	ec = &copied.Spec.EphemeralContainers[len(copied.Spec.EphemeralContainers)-1]

	return copied, ec, nil
}
