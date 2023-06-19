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
	} else {
		err := deleteCronJobIfExists(r.GetClient(), ctx, cronJob)
		if err != nil {
			return reconcile.Result{}, err
		}

		wantedSpec := getJobSpec(skipJob, job.Spec.Selector, job.Labels)

		err = ctrlutil.SetControllerReference(skipJob, &job, r.GetScheme())
		if err != nil {
			return reconcile.Result{}, err
		}

		err = r.GetClient().Get(ctx, types.NamespacedName{
			Namespace: skipJob.Namespace,
			Name:      job.Name,
		}, &job)

		if errors.IsNotFound(err) {
			util.SetCommonAnnotations(&job)

			job.Spec = wantedSpec

			err := r.GetClient().Create(ctx, &job)
			if err != nil {
				return reconcile.Result{}, err
			}

			res, err := r.UpdateStatusWithCondition(ctx, skipJob, r.GetConditionRunning(skipJob, metav1.ConditionTrue))
			if err != nil {
				return *res, err
			}
			res, err = r.UpdateStatusWithCondition(ctx, skipJob, r.GetConditionFinished(skipJob, metav1.ConditionFalse))
			if err != nil {
				return *res, err
			}
		} else if err == nil {
			if job.Status.CompletionTime == nil {
				jobPods, err := listJobPods(r.GetClient(), ctx, job)
				if err != nil {
					return reconcile.Result{}, err
				}

				// TODO Check exit code for container containing actual job, as this might fail, in which case the istio-proxy should also fail (i think)
				for _, pod := range jobPods.Items {
					if pod.Status.Phase != corev1.PodRunning {
						continue
					}

					for _, containerStatus := range pod.Status.ContainerStatuses {
						if containerStatus.Name != job.Name {
							continue
						}

						if containerStatus.State.Terminated != nil && containerStatus.State.Terminated.ExitCode == 0 {
							// If the istio-proxy is terminated we may skip all the other containers, as our work is already done
							if containerStatus.Name == "istio-proxy" {
								break
							}

							err := removeIstioProxyForPod(ctx, r.RESTClient, &pod, r.GetRestConfig(), runtime.NewParameterCodec(r.GetScheme()))
							if err != nil {
								return reconcile.Result{}, err
							}
							res, err := r.UpdateStatusWithCondition(ctx, skipJob, r.GetConditionFinished(skipJob, metav1.ConditionTrue))
							if err != nil {
								return *res, err
							}
						}
					}
				}
			} else {

			}

		} else {
			return reconcile.Result{}, err
		}
	}

	return reconcile.Result{}, nil
}

func removeIstioProxyForPod(ctx context.Context, restClient rest.Interface, pod *corev1.Pod, restConfig *rest.Config, codec runtime.ParameterCodec) error {
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
