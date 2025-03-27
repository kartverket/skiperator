package gateway

import (
	"github.com/kartverket/skiperator/v3/pkg/reconciliation"
	"github.com/kartverket/skiperator/v3/pkg/resourcegenerator/resourceutils/generator"
)

var multiGenerator = generator.NewMulti()

func Generate(r reconciliation.Reconciliation) error {
	return multiGenerator.Generate(r, "Gateway")
}
