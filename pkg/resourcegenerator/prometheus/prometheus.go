package prometheus

import (
	"github.com/kartverket/skiperator/v2/api/v1alpha1"
	"github.com/kartverket/skiperator/v2/pkg/reconciliation"
	"github.com/kartverket/skiperator/v2/pkg/resourcegenerator/resourceutils/generator"
	pov1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
)

var (
	multiGenerator        = generator.NewMulti()
	defaultScrapeInterval = pov1.Duration("60s")
)

func Generate(r reconciliation.Reconciliation) error {
	return multiGenerator.Generate(r, "PrometheusCRD")
}

func getScrapeInterval(pc *v1alpha1.PrometheusConfig) pov1.Duration {
	if pc == nil {
		return defaultScrapeInterval
	}

	return pov1.Duration(pc.ScrapeInterval)
}
