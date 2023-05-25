package applicationcontroller

import (
	"context"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/pod"
	"github.com/kartverket/skiperator/pkg/util"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
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

	if skipJob.Spec.Cron != nil {
		_, err := ctrlutil.CreateOrPatch(ctx, r.GetClient(), &cronJob, func() error {

			err := ctrlutil.SetControllerReference(skipJob, &cronJob, r.GetScheme())
			if err != nil {
				return err
			}

			util.SetCommonAnnotations(&cronJob)

			cronJob = getCronJobDefinition(skipJob)

			return nil
		})
		if err != nil {
			return reconcile.Result{}, err
		}
	} else {
		err := deleteCronJobIfExists(r.GetClient(), ctx, cronJob)
		if err != nil {
			println(err)
			return reconcile.Result{}, err
		}

		println(job.Name)
		_, err = ctrlutil.CreateOrPatch(ctx, r.GetClient(), &job, func() error {

			err := ctrlutil.SetControllerReference(skipJob, &job, r.GetScheme())
			if err != nil {
				return err
			}

			util.SetCommonAnnotations(&job)

			job = getJobDefinition(skipJob)

			return nil
		})
		if err != nil {
			println("UHHHH %s", err.Error())
			return reconcile.Result{}, err
		}
	}

	return reconcile.Result{}, nil
}

func deleteCronJobIfExists(recClient client.Client, context context.Context, cronJob batchv1.CronJob) error {
	err := recClient.Delete(context, &cronJob)
	err = client.IgnoreNotFound(err)
	if err != nil {

		return err
	}

	return nil
}

func getJobDefinition(skipJob *skiperatorv1alpha1.SKIPJob) batchv1.Job {
	return batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: skipJob.Namespace,
			Name:      skipJob.Name,
		},
		Spec: getJobSpec(skipJob),
	}
}

func getCronJobDefinition(skipJob *skiperatorv1alpha1.SKIPJob) batchv1.CronJob {
	return batchv1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      skipJob.Name,
			Namespace: skipJob.Namespace,
		},
		Spec: batchv1.CronJobSpec{
			Schedule:                skipJob.Spec.Cron.Schedule,
			TimeZone:                nil,
			StartingDeadlineSeconds: nil,
			ConcurrencyPolicy:       skipJob.Spec.Cron.ConcurrencyPolicy,
			Suspend:                 skipJob.Spec.Cron.Suspend,
			JobTemplate: batchv1.JobTemplateSpec{
				Spec: getJobSpec(skipJob),
			},
			SuccessfulJobsHistoryLimit: nil,
			FailedJobsHistoryLimit:     nil,
		},
	}
}

func getJobSpec(skipJob *skiperatorv1alpha1.SKIPJob) batchv1.JobSpec {
	jobSpec := batchv1.JobSpec{
		Parallelism:           nil,
		Completions:           nil,
		ActiveDeadlineSeconds: nil,
		PodFailurePolicy:      nil,
		BackoffLimit:          nil,
		Selector:              nil,
		ManualSelector:        nil,
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{},
			Spec:       pod.CreatePodSpec(pod.CreateJobContainer(skipJob), nil, skipJob.Name, skipJob.Spec.Container.Priority, skipJob.Spec.Container.RestartPolicy),
		},
		TTLSecondsAfterFinished: nil,
		CompletionMode:          nil,
		Suspend:                 nil,
	}

	return jobSpec
}
