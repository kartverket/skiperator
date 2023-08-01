package skipjobcontroller

import (
	"bytes"
	"context"
	"fmt"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/core"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/gcp"
	"github.com/kartverket/skiperator/pkg/util"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"strings"
	"time"
)

var (
	SKIPJobReferenceLabelKey = "skipJobOwnerName"
)

func (r *SKIPJobReconciler) reconcileJob(ctx context.Context, skipJob *skiperatorv1alpha1.SKIPJob) (reconcile.Result, error) {
	job := batchv1.Job{ObjectMeta: metav1.ObjectMeta{
		Namespace: skipJob.Namespace,
		Name:      util.ResourceNameWithHash(skipJob.Name, skipJob.Kind),
	}}

	cronJob := batchv1.CronJob{ObjectMeta: metav1.ObjectMeta{
		Namespace: skipJob.Namespace,
		Name:      util.ResourceNameWithHash(skipJob.Name, skipJob.Kind),
	}}

	gcpIdentityConfigMap, err := r.getGCPIdentityConfigMap(ctx, *skipJob)
	if err != nil {
		return reconcile.Result{}, err
	}

	if skipJob.Spec.Cron != nil {
		err := r.GetClient().Get(ctx, types.NamespacedName{
			Namespace: cronJob.Namespace,
			Name:      cronJob.Name,
		}, &cronJob)

		if errors.IsNotFound(err) {
			err := ctrlutil.SetControllerReference(skipJob, &cronJob, r.GetScheme())
			if err != nil {
				return reconcile.Result{}, err
			}

			util.SetCommonAnnotations(&cronJob)

			cronJob.Spec = getCronJobSpec(skipJob, cronJob.Name, cronJob.Spec.JobTemplate.Spec.Selector, job.Labels, gcpIdentityConfigMap)

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
			// TODO Figure out how to update CronJobs
		} else if err != nil {
			// TODO Send event due to error
			return reconcile.Result{}, err
		}
	} else {
		err := deleteCronJobIfExists(r.GetClient(), ctx, cronJob)
		if err != nil {
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
				return reconcile.Result{}, err
			}

			err = r.UpdateStatusWithCondition(ctx, skipJob, []metav1.Condition{
				r.GetConditionRunning(skipJob, metav1.ConditionTrue),
				r.GetConditionFinished(skipJob, metav1.ConditionFalse),
			})

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
			// TODO Diff current and wanted Spec for job. Not all updates are created equally, and some will force the recreation of a job.
			// Perhaps we should not allow updates to Jobs, instead forcing a recreate every time the spec differs? Doesn't really make sense to update a running job.

			jobPods := corev1.PodList{}
			err := r.GetClient().List(ctx, &jobPods, []client.ListOption{
				client.MatchingLabels{"job-name": job.Name},
			}...)
			if err != nil {
				return reconcile.Result{}, err
			}

			if len(jobPods.Items) == 0 {
				// TODO This requeue does not work. It seems like the controller ignores the requeue after
				// In the case that the job has no pods yet, we should requeue the request so that the controller can
				// check the pods when created.
				return reconcile.Result{RequeueAfter: time.Second * 5}, nil
			}

			for _, pod := range jobPods.Items {
				if pod.Status.Phase != corev1.PodRunning {
					continue
				}

				terminatedContainerStatuses := map[string]int32{}
				for _, containerStatus := range pod.Status.ContainerStatuses {
					if containerStatus.State.Terminated != nil {
						terminatedContainerStatuses[containerStatus.Name] = containerStatus.State.Terminated.ExitCode
					}
				}

				if _, exists := terminatedContainerStatuses["istio-proxy"]; exists {
					// We want to skip all further operations if the istio-proxy is terminated
					continue
				}

				if exitCode, exists := terminatedContainerStatuses[job.Labels["job-name"]]; exists {
					if exitCode == 0 {
						err := requestExitForIstioProxyContainer(ctx, r.RESTClient, &pod, r.GetRestConfig(), runtime.NewParameterCodec(r.GetScheme()))
						if err != nil {
							return reconcile.Result{}, err
						}

						// Once we know the istio-proxy pod is marked as completed, we can assume the Job is finished, and can update the status of the SKIPJob
						// immediately, so following reconciliations can use that check
						err = r.UpdateStatusWithCondition(ctx, skipJob, []metav1.Condition{
							r.GetConditionRunning(skipJob, metav1.ConditionFalse),
							r.GetConditionFinished(skipJob, metav1.ConditionTrue),
						})

						if err != nil {
							return reconcile.Result{}, err
						}
					} else {

						// TODO Operate differently if the container containing the job was not terminated successfully
					}
				} else {
					println("not finished")
					err := r.UpdateStatusWithCondition(ctx, skipJob, []metav1.Condition{
						r.GetConditionRunning(skipJob, metav1.ConditionTrue),
						r.GetConditionFinished(skipJob, metav1.ConditionFalse),
					})
					if err != nil {
						return reconcile.Result{}, err
					}
				}
			}
		} else {
			// TODO Do we need to do anything when jobs are finished?
			// Perhaps remove them after x time? CronJobs automatically clear old jobs for one
		}
	}

	return reconcile.Result{}, nil
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

func (r *SKIPJobReconciler) isSkipJobFinished(conditions []metav1.Condition) bool {
	newestSkipJobCondition, isNotEmpty := r.GetLastCondition(conditions)

	if !isNotEmpty {
		return false
	} else {
		return newestSkipJobCondition.Type == ConditionFinished && newestSkipJobCondition.Status == metav1.ConditionTrue
	}
}

func requestExitForIstioProxyContainer(ctx context.Context, restClient rest.Interface, pod *corev1.Pod, restConfig *rest.Config, codec runtime.ParameterCodec) error {
	buf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}

	// Execute into the pod to tell the istio container to quit, which in turns allows the Pod to finish
	request := restClient.
		Post().
		Namespace(pod.Namespace).
		Resource("pods").
		Name(pod.Name).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Container: "istio-proxy",
			Command:   []string{"/bin/sh", "-c", "curl --max-time 2 -s -f -XPOST http://127.0.0.1:15000/quitquitquit"},
			Stdin:     false,
			Stdout:    true,
			Stderr:    true,
			TTY:       true,
		}, codec)
	exec, err := remotecommand.NewSPDYExecutor(restConfig, "POST", request.URL())
	if err != nil {
		return fmt.Errorf("%w failed running the exec on %v/%v\n%s\n%s", err, pod.Namespace, pod.Name, buf.String(), errBuf.String())
	}
	err = exec.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdout: buf,
		Stderr: errBuf,
	})

	// Ignore container not found error, as this just means the container has been killed in this case
	if err != nil && !strings.Contains(err.Error(), "container not found") {
		return fmt.Errorf("%w failed executing on %v/%v\n%s\n%s", err, pod.Namespace, pod.Name, buf.String(), errBuf.String())
	}
	return nil
}

func deleteCronJobIfExists(recClient client.Client, context context.Context, cronJob batchv1.CronJob) error {
	err := recClient.Delete(context, &cronJob)
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
		// Not sure if we should add these fields to spec
		//TimeZone:                   nil,
		//SuccessfulJobsHistoryLimit: nil,
		//FailedJobsHistoryLimit:     nil,
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
		Parallelism:           skipJob.Spec.Job.Parallelism,
		ActiveDeadlineSeconds: skipJob.Spec.Job.ActiveDeadlineSeconds,
		BackoffLimit:          skipJob.Spec.Job.BackoffLimit,
		Template: corev1.PodTemplateSpec{
			Spec: core.CreatePodSpec(core.CreateJobContainer(skipJob, containerVolumeMounts), podVolumes, util.ResourceNameWithHash(skipJob.Name, skipJob.Kind), skipJob.Spec.Container.Priority, skipJob.Spec.Container.RestartPolicy),
		},
		TTLSecondsAfterFinished: skipJob.Spec.Job.TTLSecondsAfterFinished,
		Suspend:                 skipJob.Spec.Job.Suspend,
		// Not sure if we should add these fields to spec
		//CompletionMode: nil,
		//Completions:    nil,
	}

	// Jobs create their own selector with a random UUID. Upon creation of the Job we do not know this beforehand.
	// Therefore, simply set these again if they already exist, which would be the case if reconciling an existing job.
	if selector != nil {
		jobSpec.Selector = selector
		jobSpec.Template.ObjectMeta.Labels = labels
	}

	return jobSpec
}
