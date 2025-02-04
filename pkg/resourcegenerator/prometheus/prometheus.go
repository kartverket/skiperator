package prometheus

import (
	"fmt"

	"github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/reconciliation"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/resourceutils/generator"
	"github.com/kartverket/skiperator/pkg/util"
	pov1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	prommodel "github.com/prometheus/common/model"
)

var (
	multiGenerator        = generator.NewMulti()
	defaultScrapeInterval = pov1.Duration("60s")
	minimumInterval, _    = prommodel.ParseDuration("15s")
)

func Generate(r reconciliation.Reconciliation) error {
	return multiGenerator.Generate(r, "PrometheusCRD")
}

func getScrapeInterval(pc *v1alpha1.PrometheusConfig) (*pov1.Duration, error) {
	if pc == nil {
		return &defaultScrapeInterval, nil
	}

	interval, err := prommodel.ParseDuration(pc.ScrapeInterval)
	if err != nil {
		return nil, fmt.Errorf("failed to parse configured scrape interval %s: %w", pc.ScrapeInterval, err)
	}

	if interval < minimumInterval {
		return nil, fmt.Errorf("configured scrape interval %s is too low, must be above %s", pc.ScrapeInterval, minimumInterval)
	}

	return util.PointTo(pov1.Duration(pc.ScrapeInterval)), nil
}
