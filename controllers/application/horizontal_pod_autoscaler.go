package applicationcontroller

import (
	"context"

	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/util"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *ApplicationReconciler) reconcileHorizontalPodAutoscaler(ctx context.Context, application *skiperatorv1alpha1.Application) (reconcile.Result, error) {
	controllerName := "HorizontalPodAutoScaler"
	r.SetControllerProgressing(ctx, application, controllerName)

	horizontalPodAutoscaler := autoscalingv2.HorizontalPodAutoscaler{ObjectMeta: metav1.ObjectMeta{Namespace: application.Namespace, Name: application.Name}}
	_, err := ctrlutil.CreateOrPatch(ctx, r.GetClient(), &horizontalPodAutoscaler, func() error {
		// Set application as owner of the horizontal pod autoscaler
		err := ctrlutil.SetControllerReference(application, &horizontalPodAutoscaler, r.GetScheme())
		if err != nil {
			r.SetControllerError(ctx, application, controllerName, err)
			return err
		}

		r.SetLabelsFromApplication(ctx, &horizontalPodAutoscaler, *application)
		util.SetCommonAnnotations(&horizontalPodAutoscaler)

		horizontalPodAutoscaler.Spec.ScaleTargetRef.APIVersion = "apps/v1"
		horizontalPodAutoscaler.Spec.ScaleTargetRef.Kind = "Deployment"
		horizontalPodAutoscaler.Spec.ScaleTargetRef.Name = application.Name

		min := int32(application.Spec.Replicas.Min)
		horizontalPodAutoscaler.Spec.MinReplicas = &min
		max := int32(application.Spec.Replicas.Max)
		horizontalPodAutoscaler.Spec.MaxReplicas = max

		horizontalPodAutoscaler.Spec.Metrics = make([]autoscalingv2.MetricSpec, 1)
		horizontalPodAutoscaler.Spec.Metrics[0].Type = autoscalingv2.ResourceMetricSourceType
		horizontalPodAutoscaler.Spec.Metrics[0].Resource = &autoscalingv2.ResourceMetricSource{}
		horizontalPodAutoscaler.Spec.Metrics[0].Resource.Name = "cpu"
		horizontalPodAutoscaler.Spec.Metrics[0].Resource.Target.Type = autoscalingv2.UtilizationMetricType
		averageUtilization := int32(application.Spec.Replicas.TargetCpuUtilization)
		horizontalPodAutoscaler.Spec.Metrics[0].Resource.Target.AverageUtilization = &averageUtilization

		return nil
	})

	r.SetControllerFinishedOutcome(ctx, application, controllerName, err)

	return reconcile.Result{}, err
}
