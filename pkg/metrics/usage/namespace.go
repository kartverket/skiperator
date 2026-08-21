package usage

import (
	"github.com/prometheus/client_golang/prometheus"
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

func updateNamespace(usage clusterUsage, currentGauge *prometheus.GaugeVec) {
	counts := make(map[namespaceLabels]float64)
	for _, namespace := range usage.namespaces {
		counts[namespace]++
	}

	currentGauge.Reset()
	for key, count := range counts {
		currentGauge.With(prometheus.Labels{
			labelTeam:     key.team,
			labelDivision: key.division,
		}).Set(count)
	}
}
