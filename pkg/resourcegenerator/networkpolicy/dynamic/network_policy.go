package dynamic

import (
	"github.com/kartverket/skiperator/pkg/reconciliation"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/resourceutils/generator"
	"github.com/kartverket/skiperator/pkg/util"
)

const (
	AlloyAgentName      = "alloy"
	AlloyAgentNamespace = "grafana-alloy"
)

var multiGenerator = generator.NewMulti()

func Generate(r reconciliation.Reconciliation) error {
	if err := multiGenerator.Generate(r, "NetworkPolicy"); err != nil {
		return &util.SubResourceError{Message: "Failed to generate network policy resource", WrapErr: err, Reason: util.SubResourceGenerateFailed}
	}
	return nil
}
