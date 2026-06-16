package usage

import (
	"context"

	"github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/log"
	"github.com/prometheus/client_golang/prometheus"
	corev1 "k8s.io/api/core/v1"
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
	namespaces := &corev1.NamespaceList{}
	if err := k.List(ctx, namespaces); err != nil {
		logger.Error(err, "failed to list namespaces")
		return
	}

	counts := make(map[routingProviderMetricKey]float64)
	for _, ns := range namespaces.Items {
		team := valueOrDefault(ns.Labels[labelTeam])
		division := valueOrDefault(ns.Labels[labelDivision])

		for _, resource := range routingProviderResources {
			list := &unstructured.UnstructuredList{}
			list.SetGroupVersionKind(resource.gvr.GroupVersion().WithKind(resource.kind + "List"))

			if err := k.List(ctx, list, client.InNamespace(ns.Name)); err != nil {
				logger.Error(err, "failed to list resources for", "gvr", resource.gvr, "namespace", ns.Name)
				continue
			}

			for _, item := range list.Items {
				key := routingProviderMetricKey{
					team:            team,
					division:        division,
					kind:            resource.kind,
					routingProvider: routingProviderOrLegacy(routingProviderFromObject(item)),
				}
				counts[key]++
			}
		}
	}

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
