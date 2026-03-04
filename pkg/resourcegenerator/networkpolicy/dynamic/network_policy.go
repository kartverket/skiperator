package dynamic

import (
	"github.com/kartverket/skiperator/pkg/reconciliation"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/resourceutils/generator"
)

const (
	AlloyAgentName      = "alloy"
	AlloyAgentNamespace = "grafana-alloy"
)

var multiGenerator = generator.NewMulti()

func Generate(r reconciliation.Reconciliation) error {
	if err := multiGenerator.Generate(r, "NetworkPolicy"); err != nil {
		return err
	}
	return nil
}
