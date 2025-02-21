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

var crdGVRs = []schema.GroupVersionResource{
	v1alpha1.GroupVersion.WithResource(typeApplication),
	v1alpha1.GroupVersion.WithResource(typeSKIPJob),
	v1alpha1.GroupVersion.WithResource(typeRouting),
}

func init() {
	metadata := prometheus.GaugeOpts{
		Subsystem: metricSubsystem,
		Name:      "team_usage",
		Help:      "Number of active CRs per team and division",
	}
	labels := []string{labelTeam, labelDivision, labelType}
	registerGaugeVecFunc(metadata, labels, updateTeamCRUsage)
}

func updateTeamCRUsage(ctx context.Context, k client.Client, logger log.Logger, currentGauge *prometheus.GaugeVec) {
	// List all namespaces
	namespaces := &corev1.NamespaceList{}
	if err := k.List(ctx, namespaces); err != nil {
		logger.Error(err, "failed to list namespaces")
		return
	}

	// Initialize counts
	counts := make(map[string]map[string]float64)

	// Iterate over namespaces
	for _, ns := range namespaces.Items {
		team := ns.Labels[labelTeam]
		division := ns.Labels[labelDivision]

		key := valueOrDefault(team) + "|" + valueOrDefault(division)
		if counts[key] == nil {
			counts[key] = make(map[string]float64)
		}

		// Count instances of each CRD in the namespace
		for _, gvr := range crdGVRs {
			list := &unstructured.UnstructuredList{}
			list.SetGroupVersionKind(gvr.GroupVersion().WithKind(gvr.Resource + "List"))

			if err := k.List(ctx, list, client.InNamespace(ns.Name)); err != nil {
				logger.Error(err, "failed to list resources for", "gvr", gvr, "namespace", ns.Name)
				continue
			}

			counts[key][gvr.Resource] += float64(len(list.Items))
		}
	}

	// Reset and update metrics
	currentGauge.Reset()
	for key, metrics := range counts {
		parts := splitKey(key)
		team := parts[0]
		division := parts[1]

		for resourceType, count := range metrics {
			currentGauge.With(prometheus.Labels{
				labelTeam:     team,
				labelDivision: division,
				labelType:     resourceType,
			}).Set(count)
		}
	}
}
