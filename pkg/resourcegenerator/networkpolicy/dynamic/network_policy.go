package dynamic

import (
	"github.com/kartverket/skiperator/pkg/reconciliation"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/resourceutils/generator"
)

const (
	GrafanaAgentName      = "grafana-agent"
	GrafanaAgentNamespace = GrafanaAgentName
	AlloyAgentName        = "alloy"
	AlloyAgentNamespace   = "grafana-alloy"
)

var multiGenerator = generator.NewMulti()

func Generate(r reconciliation.Reconciliation) error {
	return multiGenerator.Generate(r, "NetworkPolicy")
}
