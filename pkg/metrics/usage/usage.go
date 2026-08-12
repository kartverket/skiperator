package usage

import (
	"context"
	"time"

	"github.com/kartverket/skiperator/pkg/log"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

const (
	metricsRefreshInterval        = 30 * time.Second
	maxApiserverOperationDuration = 20 * time.Second
)

var (
	logger  log.Logger
	kclient *client.Client
	gauges  = make(map[string]computableGauge)
)

// updateGaugeFunc is a function that will be called when it's time to
// update the gauge. It counts from the tick's clusterUsage snapshot rather than
// listing anything itself, so registering another gauge costs no API calls.
type updateGaugeFunc = func(usage clusterUsage, currentGauge *prometheus.GaugeVec)

// computableGauge is a struct that holds a gauge and a function to update
// it in order to make it easier to do bookkeeping
type computableGauge struct {
	gauge *prometheus.GaugeVec
	fn    updateGaugeFunc
}

// NewUsageMetrics initializes the usage metrics subsystem, ensuring that metrics will be updated.
// The refresh goroutine runs until ctx is cancelled, so it stops on manager shutdown / loss of leadership.
func NewUsageMetrics(ctx context.Context, k8sConfig *rest.Config, log log.Logger) error {
	if k8sConfig == nil {
		return errors.New("missing k8s REST config")
	}

	// Create a new controller-runtime client in order to
	// utilize the built-in caching mechanisms
	c, err := client.New(k8sConfig, client.Options{
		Cache: &client.CacheOptions{
			Unstructured: true,
		},
	})
	if err != nil {
		return errors.Wrap(err, "failed to create controller-runtime client")
	}

	kclient = &c
	logger = log

	// Start a background goroutine to update metrics periodically
	go func() {
		// initial update
		initialCtx, initialCancel := apiserverCtx()
		updateMetrics(initialCtx)
		initialCancel()
		// regular update
		ticker := time.NewTicker(metricsRefreshInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				tickCtx, cancel := apiserverCtx()
				updateMetrics(tickCtx)
				cancel()
			}
		}
	}()

	return nil
}

func updateMetrics(ctx context.Context) {
	logger.Debug("refreshing data and updating metrics")
	usage, ok := collectClusterUsage(ctx, *kclient, logger)
	if !ok {
		return
	}
	for _, g := range gauges {
		g.fn(usage, g.gauge)
	}
}

// registerGaugeVecFunc registers a gauge with the given options and label names with the
// shared metrics registry.
func registerGaugeVecFunc(opts prometheus.GaugeOpts, labelNames []string, fn updateGaugeFunc) {
	g := prometheus.NewGaugeVec(opts, labelNames)
	metrics.Registry.MustRegister(g)
	gauges[opts.Name] = computableGauge{gauge: g, fn: fn}
}

// Ensure empty string if label is missing to avoid "no metric for label set" errors
func valueOrDefault(value string) string {
	if len(value) == 0 {
		return unknownValue
	}
	return value
}

func apiserverCtx() (context.Context, context.CancelFunc) {
	b := context.Background()
	return context.WithTimeout(b, maxApiserverOperationDuration)
}
