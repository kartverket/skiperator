package usage

import (
	"github.com/prometheus/client_golang/prometheus"
)

type teamMetricKey struct {
	team     string
	division string
	kind     string
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

func updateTeamCRUsage(usage clusterUsage, currentGauge *prometheus.GaugeVec) {
	counts := make(map[teamMetricKey]float64)
	for _, resource := range usage.resources {
		counts[teamMetricKey{team: resource.team, division: resource.division, kind: resource.kind}]++
	}

	currentGauge.Reset()
	for key, count := range counts {
		currentGauge.With(prometheus.Labels{
			labelTeam:     key.team,
			labelDivision: key.division,
			labelType:     key.kind,
		}).Set(count)
	}
}
