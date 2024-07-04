package hpa

import (
	"context"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/log"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/resourceutils"
	"github.com/kartverket/skiperator/pkg/util"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Generate(ctx context.Context, application *skiperatorv1alpha1.Application) (*autoscalingv2.HorizontalPodAutoscaler, error) {
	ctxLog := log.FromContext(ctx)
	ctxLog.Debug("Attempting to generate ingress gateways for application", application.Name)

	horizontalPodAutoscaler := autoscalingv2.HorizontalPodAutoscaler{ObjectMeta: metav1.ObjectMeta{Namespace: application.Namespace, Name: application.Name}}

	resourceutils.SetApplicationLabels(&horizontalPodAutoscaler, application)
	resourceutils.SetCommonAnnotations(&horizontalPodAutoscaler)

	replicas, err := skiperatorv1alpha1.GetScalingReplicas(application.Spec.Replicas)
	if err != nil {
		return nil, err
	}

	horizontalPodAutoscaler.Spec = autoscalingv2.HorizontalPodAutoscalerSpec{
		ScaleTargetRef: autoscalingv2.CrossVersionObjectReference{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
			Name:       application.Name,
		},
		MinReplicas: util.PointTo(int32(replicas.Min)),
		MaxReplicas: int32(replicas.Max),
		Metrics: []autoscalingv2.MetricSpec{
			{
				Type: autoscalingv2.ResourceMetricSourceType,
				Resource: &autoscalingv2.ResourceMetricSource{
					Name: "cpu",
					Target: autoscalingv2.MetricTarget{
						Type:               autoscalingv2.UtilizationMetricType,
						AverageUtilization: util.PointTo(int32(replicas.TargetCpuUtilization)),
					},
				},
			},
		},
	}

	return &horizontalPodAutoscaler, nil
}
