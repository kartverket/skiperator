package telemetry

import (
	"fmt"
	"github.com/kartverket/skiperator/pkg/reconciliation"
	"github.com/kartverket/skiperator/pkg/util"
	"google.golang.org/protobuf/types/known/wrapperspb"
	telemetryapiv1 "istio.io/api/telemetry/v1"
	"istio.io/api/type/v1beta1"
	telemetryv1 "istio.io/client-go/pkg/apis/telemetry/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Generate(r reconciliation.Reconciliation) error {

	ctxLog := r.GetLogger()
	ctxLog.Debug("Attempting to generate istio telemetry resource for skipobj", "skipobj", r.GetSKIPObject().GetName())

	if r.GetType() != reconciliation.ApplicationType && r.GetType() != reconciliation.JobType {
		return fmt.Errorf("istio telemetry resource only supports the application and skipjob type")
	}

	object := r.GetSKIPObject()

	istioSettings := object.GetCommonSpec().IstioSettings

	// TODO: Should we add KindPostFixedName() for the SKIPObject interface and use that for the telemetry name to avoid conflicts?
	telemetry := telemetryv1.Telemetry{ObjectMeta: metav1.ObjectMeta{Namespace: object.GetNamespace(), Name: object.GetName()}}

	var telemetryTracing []*telemetryapiv1.Tracing
	for _, tracingSetting := range istioSettings.Telemetry.Tracing {
		telemetryTracing = append(telemetryTracing, &telemetryapiv1.Tracing{
			Providers: []*telemetryapiv1.ProviderRef{
				{
					Name: util.IstioTraceProvider,
				},
			},
			RandomSamplingPercentage: util.PointTo(wrapperspb.DoubleValue{
				Value: float64(tracingSetting.RandomSamplingPercentage),
			}),
		})
	}
	telemetry.Spec = telemetryapiv1.Telemetry{
		Tracing: telemetryTracing,
		Selector: &v1beta1.WorkloadSelector{
			MatchLabels: object.GetDefaultLabels(),
		},
	}

	r.AddResource(&telemetry)

	ctxLog.Debug("Finished generating telemetry for skipobj", "skipobj", r.GetSKIPObject().GetName())
	return nil
}
