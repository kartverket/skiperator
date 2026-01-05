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
	[]string{"name", "namespace", "kind", ignoreLabel},
)

func init() {
	metrics.Registry.MustRegister(ignoredResourceGauge)
}
 
func ExposeIgnoreResource(obj client.Object, v float64){
    labels := obj.GetLabels()
    kind := obj.GetObjectKind().GroupVersionKind().Kind
    ignoredResourceGauge.WithLabelValues(
	    obj.GetName(), obj.GetNamespace(), kind, labels[ignoreLabel],
	).Set(v)
}