package certificate

import (
	"github.com/kartverket/skiperator/pkg/reconciliation"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/resourceutils/generator"
	"github.com/kartverket/skiperator/pkg/util"
)

const (
	IstioGatewayNamespace = "istio-gateways"
)

var multiGenerator = generator.NewMulti()

func Generate(r reconciliation.Reconciliation) error {
	if err := multiGenerator.Generate(r, "Certificate"); err != nil {
		return &util.SubResourceError{Message: "Failed to generate certificate resource", WrapErr: err, Reason: util.SubResourceGenerateFailed}
	}
	return nil
}
