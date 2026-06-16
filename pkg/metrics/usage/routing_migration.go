package usage

import (
	"context"

	commontypes "github.com/kartverket/skiperator/api/common"
	"github.com/kartverket/skiperator/pkg/log"
	"github.com/prometheus/client_golang/prometheus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
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

func updateRoutingMigrationStalled(ctx context.Context, k client.Client, logger log.Logger, currentGauge *prometheus.GaugeVec) {
	namespaces := &corev1.NamespaceList{}
	if err := k.List(ctx, namespaces); err != nil {
		logger.Error(err, "failed to list namespaces")
		return
	}

	counts := make(map[routingMigrationMetricKey]float64)
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
				if !hasStalledRoutingMigration(item) {
					continue
				}
				key := routingMigrationMetricKey{
					team:     team,
					division: division,
					kind:     resource.kind,
				}
				counts[key]++
			}
		}
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
		if conditionType == commontypes.StandardRoutingReadyConditionType && reason == migrationStalledReason {
			return true
		}
	}
	return false
}
