package controllers

import (
	"context"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	autoscalingv2beta2 "k8s.io/api/autoscaling/v2beta2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

//+kubebuilder:rbac:groups=skiperator.kartverket.no,resources=applications,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=autoscaling,resources=horizontalpodautoscalers,verbs=get;list;watch;create;update;patch;delete

type HorizontalPodAutoscalerReconciler struct {
	client client.Client
	scheme *runtime.Scheme
}

func (r *HorizontalPodAutoscalerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.client = mgr.GetClient()
	r.scheme = mgr.GetScheme()

	return newControllerManagedBy[*skiperatorv1alpha1.Application](mgr).
		For(&skiperatorv1alpha1.Application{}).
		Owns(&autoscalingv2beta2.HorizontalPodAutoscaler{}).
		Complete(r)
}

func (r *HorizontalPodAutoscalerReconciler) Reconcile(ctx context.Context, application *skiperatorv1alpha1.Application) (reconcile.Result, error) {
	application.FillDefaults()

	horizontalPodAutoscaler := autoscalingv2beta2.HorizontalPodAutoscaler{ObjectMeta: metav1.ObjectMeta{Namespace: application.Namespace, Name: application.Name}}
	_, err := ctrlutil.CreateOrPatch(ctx, r.client, &horizontalPodAutoscaler, func() error {
		// Set application as owner of the horizontal pod autoscaler
		err := ctrlutil.SetControllerReference(application, &horizontalPodAutoscaler, r.scheme)
		if err != nil {
			return err
		}

		horizontalPodAutoscaler.Spec.ScaleTargetRef.APIVersion = "apps/v1"
		horizontalPodAutoscaler.Spec.ScaleTargetRef.Kind = "Deployment"
		horizontalPodAutoscaler.Spec.ScaleTargetRef.Name = application.Name

		min := int32(application.Spec.Replicas.Min)
		horizontalPodAutoscaler.Spec.MinReplicas = &min
		max := int32(application.Spec.Replicas.Max)
		horizontalPodAutoscaler.Spec.MaxReplicas = max

		horizontalPodAutoscaler.Spec.Metrics = make([]autoscalingv2beta2.MetricSpec, 1)
		horizontalPodAutoscaler.Spec.Metrics[0].Type = autoscalingv2beta2.ResourceMetricSourceType
		horizontalPodAutoscaler.Spec.Metrics[0].Resource = &autoscalingv2beta2.ResourceMetricSource{}
		horizontalPodAutoscaler.Spec.Metrics[0].Resource.Name = "cpu"
		horizontalPodAutoscaler.Spec.Metrics[0].Resource.Target.Type = autoscalingv2beta2.UtilizationMetricType
		averageUtilization := int32(application.Spec.Replicas.TargetCpuUtilization)
		horizontalPodAutoscaler.Spec.Metrics[0].Resource.Target.AverageUtilization = &averageUtilization

		return nil
	})
	return reconcile.Result{}, err
}
