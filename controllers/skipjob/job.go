package applicationcontroller

import (
	"bytes"
	"context"
	"fmt"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/core"
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

	if skipJob.Spec.Cron != nil {
		// TODO CRUD CronJobs without CreateOrPatch as well
		// TODO Handle istio-proxy for CronJobs as well
		_, err := ctrlutil.CreateOrPatch(ctx, r.GetClient(), &cronJob, func() error {

			err := ctrlutil.SetControllerReference(skipJob, &cronJob, r.GetScheme())
			if err != nil {
				return err
			}

			util.SetCommonAnnotations(&cronJob)

			cronJob.Spec = getCronJobSpec(skipJob, cronJob.Spec.JobTemplate.Spec.Selector, job.Labels)

			return nil
		})
		if err != nil {
			return reconcile.Result{}, err
		}

		// Requeue the reconciliation after creating the Cron Job to allow the control plane to create subresources
		return reconcile.Result{
			Requeue:      true,
			RequeueAfter: 5,
		}, nil
	} else {
		err := deleteCronJobIfExists(r.GetClient(), ctx, cronJob)
		if err != nil {
			return reconcile.Result{}, err
		}
	}

	err := r.GetClient().Get(ctx, types.NamespacedName{
		Namespace: skipJob.Namespace,
		Name:      job.Name,
	}, &job)

	if errors.IsNotFound(err) {
		util.SetCommonAnnotations(&job)

		err = ctrlutil.SetControllerReference(skipJob, &job, r.GetScheme())
		if err != nil {
			return reconcile.Result{}, err
		}

		wantedSpec := getJobSpec(skipJob, job.Spec.Selector, job.Labels)
		job.Spec = wantedSpec

		err := r.GetClient().Create(ctx, &job)
		if err != nil {
			return reconcile.Result{}, err
		}

		res, err := r.UpdateStatusWithCondition(ctx, skipJob, []metav1.Condition{
			r.GetConditionRunning(skipJob, metav1.ConditionTrue),
			r.GetConditionFinished(skipJob, metav1.ConditionFalse),
		})
		if err != nil {
			return *res, err
		}
	} else if err == nil {
		if !r.isSkipJobFinished(skipJob.Status.Conditions) && job.Status.CompletionTime == nil {

			// TODO Diff current and wanted Spec for job. Not all updates are created equally, and some will force the recreation of a job.
			// Perhaps we should not allow updates to Jobs, instead forcing a recreate every time the spec differs? Doesn't really make sense to update a running job.

			jobPods, err := listJobPods(r.GetClient(), ctx, job)
			if err != nil {
				return reconcile.Result{}, err
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

				if exitCode, exists := terminatedContainerStatuses[job.Name]; exists {
					if exitCode == 0 {
						err := requestExitForIstioProxyContainer(ctx, r.RESTClient, &pod, r.GetRestConfig(), runtime.NewParameterCodec(r.GetScheme()))
						if err != nil {
							return reconcile.Result{}, err
						}

						// Once we know the istio-proxy pod is marked as completed, we can assume the Job is finished, and can update the status of the SKIPJob
						// immediately, so following reconciliations can use that check
						res, err := r.UpdateStatusWithCondition(ctx, skipJob, []metav1.Condition{
							r.GetConditionFinished(skipJob, metav1.ConditionTrue),
						})
						if err != nil {
							return *res, err
						}
					} else {
						// TODO Operate differently if the container containing the job was not terminated successfully
					}
				}
			}
		} else {
			// TODO Do we need to do anything when jobs are finished?
		}

	} else {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
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
	if err != nil {
		return fmt.Errorf("%w failed executing on %v/%v\n%s\n%s", err, pod.Namespace, pod.Name, buf.String(), errBuf.String())
	}
	return nil
}

func listJobPods(recClient client.Client, ctx context.Context, job batchv1.Job) (corev1.PodList, error) {
	jobPods := corev1.PodList{}
	err := recClient.List(ctx, &jobPods, []client.ListOption{
		client.InNamespace(job.Namespace),
		client.MatchingLabels{"job-name": job.Labels["job-name"]},
	}...)

	return jobPods, err
}

func deleteCronJobIfExists(recClient client.Client, context context.Context, cronJob batchv1.CronJob) error {
	err := recClient.Delete(context, &cronJob)
	err = client.IgnoreNotFound(err)
	if err != nil {

		return err
	}

	return nil
}

func getCronJobSpec(skipJob *skiperatorv1alpha1.SKIPJob, selector *metav1.LabelSelector, labels map[string]string) batchv1.CronJobSpec {
	return batchv1.CronJobSpec{
		Schedule:                skipJob.Spec.Cron.Schedule,
		StartingDeadlineSeconds: skipJob.Spec.Cron.StartingDeadlineSeconds,
		ConcurrencyPolicy:       skipJob.Spec.Cron.ConcurrencyPolicy,
		Suspend:                 skipJob.Spec.Cron.Suspend,
		JobTemplate: batchv1.JobTemplateSpec{
			Spec: getJobSpec(skipJob, selector, labels),
		},
		// Not sure if we should add these fields to spec
		//TimeZone:                   nil,
		//SuccessfulJobsHistoryLimit: nil,
		//FailedJobsHistoryLimit:     nil,
	}
}

func getJobSpec(skipJob *skiperatorv1alpha1.SKIPJob, selector *metav1.LabelSelector, labels map[string]string) batchv1.JobSpec {

	jobSpec := batchv1.JobSpec{
		Parallelism:           skipJob.Spec.Job.Parallelism,
		ActiveDeadlineSeconds: skipJob.Spec.Job.ActiveDeadlineSeconds,
		BackoffLimit:          skipJob.Spec.Job.BackoffLimit,
		Template: corev1.PodTemplateSpec{
			Spec: core.CreatePodSpec(core.CreateJobContainer(skipJob), nil, util.ResourceNameWithHash(skipJob.Name, skipJob.Kind), skipJob.Spec.Container.Priority, skipJob.Spec.Container.RestartPolicy),
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
