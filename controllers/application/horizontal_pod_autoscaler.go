package applicationcontroller

import (
	"context"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/util"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *ApplicationReconciler) reconcileHorizontalPodAutoscaler(ctx context.Context, application *skiperatorv1alpha1.Application) (reconcile.Result, error) {
	controllerName := "HorizontalPodAutoScaler"
	r.SetControllerProgressing(ctx, application, controllerName)

	horizontalPodAutoscaler := autoscalingv2.HorizontalPodAutoscaler{ObjectMeta: metav1.ObjectMeta{Namespace: application.Namespace, Name: application.Name}}
	if shouldScaleToZero(application.Spec.Replicas.Min, application.Spec.Replicas.Max) {
		err := r.GetClient().Delete(ctx, &horizontalPodAutoscaler)
		err = client.IgnoreNotFound(err)
		if err != nil {
			r.SetControllerError(ctx, application, controllerName, err)
			return reconcile.Result{}, err
		}
		r.SetControllerFinishedOutcome(ctx, application, controllerName, nil)
		return reconcile.Result{}, nil
	}

	_, err := ctrlutil.CreateOrPatch(ctx, r.GetClient(), &horizontalPodAutoscaler, func() error {
		// Set application as owner of the horizontal pod autoscaler
		err := ctrlutil.SetControllerReference(application, &horizontalPodAutoscaler, r.GetScheme())
		if err != nil {
			r.SetControllerError(ctx, application, controllerName, err)
			return err
		}

		r.SetLabelsFromApplication(&horizontalPodAutoscaler, *application)
		util.SetCommonAnnotations(&horizontalPodAutoscaler)

		horizontalPodAutoscaler.Spec = autoscalingv2.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: autoscalingv2.CrossVersionObjectReference{
				APIVersion: "apps/v1",
				Kind:       "Deployment",
				Name:       application.Name,
			},
			MinReplicas: util.PointTo(int32(application.Spec.Replicas.Min)),
			MaxReplicas: int32(application.Spec.Replicas.Max),
			Metrics: []autoscalingv2.MetricSpec{
				{
					Type: autoscalingv2.ResourceMetricSourceType,
					Resource: &autoscalingv2.ResourceMetricSource{
						Name: "cpu",
						Target: autoscalingv2.MetricTarget{
							Type:               autoscalingv2.UtilizationMetricType,
							AverageUtilization: util.PointTo(int32(application.Spec.Replicas.TargetCpuUtilization)),
						},
					},
				},
			},
		}

		return nil
	})

	r.SetControllerFinishedOutcome(ctx, application, controllerName, err)

	return reconcile.Result{}, err
}
