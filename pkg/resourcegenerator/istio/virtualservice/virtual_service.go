package virtualservice

import (
	"github.com/kartverket/skiperator/pkg/reconciliation"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/resourceutils/generator"
)

var multiGenerator = generator.NewMulti()

func Generate(r reconciliation.Reconciliation) error {
	err := multiGenerator.Generate(r, "VirtualService")
	if err != nil {
		return err
	}
	return nil
}
