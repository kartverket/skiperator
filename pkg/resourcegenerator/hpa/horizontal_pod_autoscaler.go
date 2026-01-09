package hpa

import (
	"fmt"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/v2/api/v1alpha1"
	"github.com/kartverket/skiperator/v2/pkg/reconciliation"
	"github.com/kartverket/skiperator/v2/pkg/resourcegenerator/resourceutils"
	"github.com/kartverket/skiperator/v2/pkg/util"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Generate(r reconciliation.Reconciliation) error {
	ctxLog := r.GetLogger()
	if r.GetType() != reconciliation.ApplicationType {
		return fmt.Errorf("unsupported type %s in horizontal pod autoscaler", r.GetType())
	}
	application, ok := r.GetSKIPObject().(*skiperatorv1alpha1.Application)
	if !ok {
		err := fmt.Errorf("failed to cast resource to application")
		ctxLog.Error(err, "Failed to generate horizontal pod autoscaler")
		return err
	}

	ctxLog.Debug("Attempting to generate HPA for application", "application", application.Name)

	if resourceutils.ShouldScaleToZero(application.Spec.Replicas) || !skiperatorv1alpha1.IsHPAEnabled(application.Spec.Replicas) {
		ctxLog.Debug("Skipping horizontal pod autoscaler generation for application")
		return nil
	}

	horizontalPodAutoscaler := autoscalingv2.HorizontalPodAutoscaler{ObjectMeta: metav1.ObjectMeta{Namespace: application.Namespace, Name: application.Name}}

	replicas, err := skiperatorv1alpha1.GetScalingReplicas(application.Spec.Replicas)
	if err != nil {
		return err
	}

	metrics := []autoscalingv2.MetricSpec{}
	if replicas.TargetCpuUtilization != 0 {
		metrics = append(metrics, autoscalingv2.MetricSpec{
			Type: autoscalingv2.ResourceMetricSourceType,
			Resource: &autoscalingv2.ResourceMetricSource{
				Name: "cpu",
				Target: autoscalingv2.MetricTarget{
					Type:               autoscalingv2.UtilizationMetricType,
					AverageUtilization: util.PointTo(int32(replicas.TargetCpuUtilization)),
				},
			},
		})
	}
	if replicas.TargetMemoryUtilization != 0 {
		metrics = append(metrics, autoscalingv2.MetricSpec{
			Type: autoscalingv2.ResourceMetricSourceType,
			Resource: &autoscalingv2.ResourceMetricSource{
				Name: "memory",
				Target: autoscalingv2.MetricTarget{
					Type:               autoscalingv2.UtilizationMetricType,
					AverageUtilization: util.PointTo(int32(replicas.TargetMemoryUtilization)),
				},
			},
		})
	}

	horizontalPodAutoscaler.Spec = autoscalingv2.HorizontalPodAutoscalerSpec{
		ScaleTargetRef: autoscalingv2.CrossVersionObjectReference{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
			Name:       application.Name,
		},
		MinReplicas: util.PointTo(int32(replicas.Min)),
		MaxReplicas: int32(replicas.Max),
		Metrics:     metrics,
	}
	

	r.AddResource(&horizontalPodAutoscaler)

	return nil
}
