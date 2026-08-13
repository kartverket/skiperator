package usage

import (
	"github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type routingProviderMetricKey struct {
	team            string
	division        string
	kind            string
	routingProvider string
}

func init() {
	metadata := prometheus.GaugeOpts{
		Subsystem: metricSubsystem,
		Name:      "routing_provider_usage",
		Help:      "Number of active routable CRs by routing provider",
	}
	labels := []string{labelTeam, labelDivision, labelType, labelRoutingProvider}
	registerGaugeVecFunc(metadata, labels, updateRoutingProviderUsage)
}

func updateRoutingProviderUsage(usage clusterUsage, currentGauge *prometheus.GaugeVec) {
	counts := make(map[routingProviderMetricKey]float64)
	usage.routables(func(resource usageResource) {
		counts[routingProviderMetricKey{
			team:            resource.team,
			division:        resource.division,
			kind:            resource.kind,
			routingProvider: routingProviderFromObject(resource.item),
		}]++
	})

	currentGauge.Reset()
	for key, count := range counts {
		currentGauge.With(prometheus.Labels{
			labelTeam:            key.team,
			labelDivision:        key.division,
			labelType:            key.kind,
			labelRoutingProvider: key.routingProvider,
		}).Set(count)
	}
}

// routingProviderFromObject reads spec.routingProvider, defaulting to Legacy for
// objects written before the field existed.
func routingProviderFromObject(obj unstructured.Unstructured) string {
	provider, _, _ := unstructured.NestedString(obj.Object, "spec", "routingProvider")
	if provider == "" {
		return string(v1alpha1.RoutingProviderLegacy)
	}
	return provider
}
