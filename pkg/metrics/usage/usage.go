package usage

import (
	"context"
	"strings"
	"time"

	"github.com/kartverket/skiperator/pkg/log"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
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
// update the gauge.
type updateGaugeFunc = func(ctx context.Context, k client.Client, logger log.Logger, currentGauge *prometheus.GaugeVec)

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

// nsLabels holds the organizational labels copied from a namespace.
type nsLabels struct {
	team     string
	division string
}

// forEachRoutableResource lists every Application and Routing in the cluster
// once (not per-namespace) and invokes fn with each item plus its namespace's
// team/division labels. This replaces an O(namespaces*kinds) per-namespace List
// fan-out with one List per kind.
func forEachRoutableResource(ctx context.Context, k client.Client, logger log.Logger, fn func(item unstructured.Unstructured, kind, team, division string)) {
	namespaces := &corev1.NamespaceList{}
	if err := k.List(ctx, namespaces); err != nil {
		logger.Error(err, "failed to list namespaces")
		return
	}
	labelsByNamespace := make(map[string]nsLabels, len(namespaces.Items))
	for _, ns := range namespaces.Items {
		labelsByNamespace[ns.Name] = nsLabels{
			team:     valueOrDefault(ns.Labels[labelTeam]),
			division: valueOrDefault(ns.Labels[labelDivision]),
		}
	}

	for _, resource := range routingProviderResources {
		list := &unstructured.UnstructuredList{}
		list.SetGroupVersionKind(resource.gvr.GroupVersion().WithKind(resource.kind + "List"))
		if err := k.List(ctx, list); err != nil {
			logger.Error(err, "failed to list resources", "gvr", resource.gvr)
			continue
		}
		for _, item := range list.Items {
			labels, ok := labelsByNamespace[item.GetNamespace()]
			if !ok {
				// Namespace appeared between the namespace and resource list calls;
				// fall back to defaults instead of empty label values.
				labels = nsLabels{team: unknownValue, division: unknownValue}
			}
			fn(item, resource.kind, labels.team, labels.division)
		}
	}
}

func updateMetrics(ctx context.Context) {
	logger.Debug("refreshing data and updating metrics")
	for _, g := range gauges {
		g.fn(ctx, *kclient, logger, g.gauge)
	}
}

// registerGaugeVecFunc registers a gauge with the given options and label names with the
// shared metrics registry.
func registerGaugeVecFunc(opts prometheus.GaugeOpts, labelNames []string, fn updateGaugeFunc) {
	g := prometheus.NewGaugeVec(opts, labelNames)
	metrics.Registry.MustRegister(g)
	gauges[opts.Name] = computableGauge{gauge: g, fn: fn}
}

// Helper function to split key back into label values
func splitKey(key string) []string {
	parts := [2]string{unknownValue, unknownValue}
	if key == "" {
		return parts[:]
	}

	split := strings.SplitN(key, "|", 2)
	copy(parts[:], split)
	return parts[:]
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
