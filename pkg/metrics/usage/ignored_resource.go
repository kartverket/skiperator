package usage

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var ignoredResourceGauge = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Subsystem: metricSubsystem,
		Name:      "ignored_resource",
		Help:      "Resource ignored by Skiperator",
	},
	[]string{"name", "namespace", "kind"},
)

func init() {
	metrics.Registry.MustRegister(ignoredResourceGauge)
}

func ExposeIgnoredResource(obj client.Object) {
	kind := obj.GetObjectKind().GroupVersionKind().Kind
	res := ignoredResourceGauge.WithLabelValues(
		obj.GetName(), obj.GetNamespace(), kind,
	)

	res.Set(1)
}

func RemoveIgnoredResource(obj client.Object) {
	ignoredResourceGauge.Delete(
		ignoreLabels(obj.GetName(), obj.GetNamespace(), obj.GetObjectKind().GroupVersionKind().Kind),
	)
}

func ignoreLabels(name string, namespace string, kind string) prometheus.Labels {
	labels := make(prometheus.Labels)
	labels["name"] = name
	labels["namespace"] = namespace
	labels["kind"] = kind

	return labels
}
