package gateway

import (
	"github.com/kartverket/skiperator/pkg/reconciliation"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/resourceutils/generator"
	"github.com/kartverket/skiperator/pkg/util"
)

var multiGenerator = generator.NewMulti()

func Generate(r reconciliation.Reconciliation) error {
	if err := multiGenerator.Generate(r, "Gateway"); err != nil {
		return &util.SubResourceError{Message: "Failed to generate gateway resource", WrapErr: err, Reason: util.SubResourceGenerateFailed}
	}
	return nil
}
