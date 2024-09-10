package telemetry

import (
	"fmt"
	"github.com/kartverket/skiperator/pkg/reconciliation"
	"github.com/kartverket/skiperator/pkg/util"
	"google.golang.org/protobuf/types/known/wrapperspb"
	telemetryapiv1 "istio.io/api/telemetry/v1"
	telemetryv1 "istio.io/client-go/pkg/apis/telemetry/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Generate(r reconciliation.Reconciliation) error {

	ctxLog := r.GetLogger()
	ctxLog.Debug("Attempting to generate istio telemetry resource for skipobj", "skipobj", r.GetSKIPObject().GetName())

	if r.GetType() != reconciliation.ApplicationType {
		return fmt.Errorf("istio sidecar resource only supports the namespace type")
	}

	object := r.GetSKIPObject()

	istioSettings := object.GetCommonSpec().IstioSettings

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
	}

	r.AddResource(&telemetry)

	ctxLog.Debug("Finished generating telemetry for skipobj", "skipobj", r.GetSKIPObject().GetName())
	return nil
}
