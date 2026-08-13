package usage

import (
	commontypes "github.com/kartverket/skiperator/api/common"
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type routingMigrationMetricKey struct {
	team     string
	division string
	kind     string
}

func init() {
	metadata := prometheus.GaugeOpts{
		Subsystem: metricSubsystem,
		Name:      "routing_migration_stalled",
		Help:      "Number of active routable CRs with stalled Gateway API migration",
	}
	labels := []string{labelTeam, labelDivision, labelType}
	registerGaugeVecFunc(metadata, labels, updateRoutingMigrationStalled)
}

func updateRoutingMigrationStalled(usage clusterUsage, currentGauge *prometheus.GaugeVec) {
	counts := make(map[routingMigrationMetricKey]float64)
	usage.routables(func(resource usageResource) {
		if !hasStalledRoutingMigration(resource.item) {
			return
		}
		counts[routingMigrationMetricKey{team: resource.team, division: resource.division, kind: resource.kind}]++
	})

	currentGauge.Reset()
	for key, count := range counts {
		currentGauge.With(prometheus.Labels{
			labelTeam:     key.team,
			labelDivision: key.division,
			labelType:     key.kind,
		}).Set(count)
	}
}

func hasStalledRoutingMigration(obj unstructured.Unstructured) bool {
	conditions, _, _ := unstructured.NestedSlice(obj.Object, "status", "conditions")
	for _, condition := range conditions {
		conditionMap, ok := condition.(map[string]interface{})
		if !ok {
			continue
		}
		conditionType, _, _ := unstructured.NestedString(conditionMap, "type")
		reason, _, _ := unstructured.NestedString(conditionMap, "reason")
		// Ready can also use MigrationStalled. Count only the Gateway API
		// routing condition so the metric tracks migration health specifically.
		if conditionType == commontypes.StandardRoutingReadyConditionType && reason == commontypes.MigrationStalledReason {
			return true
		}
	}
	return false
}
