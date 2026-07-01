package usage

import (
	"context"

	"github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/log"
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type routingProviderResource struct {
	gvr  schema.GroupVersionResource
	kind string
}

type routingProviderMetricKey struct {
	team            string
	division        string
	kind            string
	routingProvider string
}

var routingProviderResources = []routingProviderResource{
	{gvr: v1alpha1.GroupVersion.WithResource("applications"), kind: typeApplication},
	{gvr: v1alpha1.GroupVersion.WithResource("routings"), kind: typeRouting},
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

func updateRoutingProviderUsage(ctx context.Context, k client.Client, logger log.Logger, currentGauge *prometheus.GaugeVec) {
	counts := make(map[routingProviderMetricKey]float64)
	forEachRoutableResource(ctx, k, logger, func(item unstructured.Unstructured, kind, team, division string) {
		key := routingProviderMetricKey{
			team:            team,
			division:        division,
			kind:            kind,
			routingProvider: routingProviderOrLegacy(routingProviderFromObject(item)),
		}
		counts[key]++
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

func routingProviderFromObject(obj unstructured.Unstructured) string {
	provider, _, _ := unstructured.NestedString(obj.Object, "spec", "routingProvider")
	return provider
}

func routingProviderOrLegacy(provider string) string {
	if provider == "" {
		return string(v1alpha1.RoutingProviderLegacy)
	}

	return provider
}
