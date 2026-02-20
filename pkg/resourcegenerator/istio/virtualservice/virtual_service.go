package virtualservice

import (
	"github.com/kartverket/skiperator/pkg/reconciliation"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/resourceutils/generator"
	"github.com/kartverket/skiperator/pkg/util"
)

var multiGenerator = generator.NewMulti()

func Generate(r reconciliation.Reconciliation) error {
	err := multiGenerator.Generate(r, "VirtualService")
	if err != nil {
		return &util.SubResourceError{Message: "Failed to generate virtual service resource", WrapErr: err, Reason: util.SubResourceGenerateFailed}
	}
	return nil
}
