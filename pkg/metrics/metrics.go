package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	ReconcileFinished = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name:      "reconcile_finished_total",
		Namespace: "skiperator",
		Help:      "number of reconciliation loops finished in total",
	}, []string{"name", "namespace", "controller_part"})

	ReconcileFailed = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name:      "reconcile_failed_total",
		Namespace: "skiperator",
		Help:      "number of reconciliation loops failed in total",
	}, []string{"app_name", "app_namespace", "controller_part"})
)

func Register(registry prometheus.Registerer) {
	registry.MustRegister(
		ReconcileFinished,
		ReconcileFailed,
	)
}
