package dynamic

import (
	"github.com/kartverket/skiperator/v2/pkg/reconciliation"
	"github.com/kartverket/skiperator/v2/pkg/resourcegenerator/resourceutils/generator"
)

const (
	AlloyAgentName      = "alloy"
	AlloyAgentNamespace = "grafana-alloy"
)

var multiGenerator = generator.NewMulti()

func Generate(r reconciliation.Reconciliation) error {
	return multiGenerator.Generate(r, "NetworkPolicy")
}
