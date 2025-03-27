package usage

import (
	"context"

	"github.com/kartverket/skiperator/v3/pkg/log"
	"github.com/prometheus/client_golang/prometheus"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func init() {
	metadata := prometheus.GaugeOpts{
		Subsystem: metricSubsystem,
		Name:      "namespace_metadata",
		Help:      "Metadata regarding number of namespaces per team and division",
	}
	labels := []string{labelTeam, labelDivision}
	registerGaugeVecFunc(metadata, labels, updateNamespace)
}

func updateNamespace(ctx context.Context, k client.Client, logger log.Logger, currentGauge *prometheus.GaugeVec) {
	namespaces := &corev1.NamespaceList{}
	if err := k.List(ctx, namespaces); err != nil {
		logger.Error(err, "failed to list namespaces")
		return
	}

	counts := make(map[string]map[string]float64)

	// Count namespaces grouped by label combination
	for _, ns := range namespaces.Items {
		team := ns.Labels[labelTeam]
		division := ns.Labels[labelDivision]

		key := valueOrDefault(team) + "|" + valueOrDefault(division)
		if counts[key] == nil {
			counts[key] = map[string]float64{}
		}
		counts[key][countKey]++ // Increment count
	}

	// Reset and update metrics
	currentGauge.Reset()
	for key, _ := range counts {
		labels := map[string]string{}
		parts := splitKey(key)
		labels[labelTeam] = parts[0]
		labels[labelDivision] = parts[1]

		currentGauge.With(labels).Set(counts[key][countKey])
	}
}
