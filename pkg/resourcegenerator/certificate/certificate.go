package certificate

import (
	"github.com/kartverket/skiperator/pkg/reconciliation"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/resourceutils/generator"
)

const (
	IstioGatewayNamespace = "istio-gateways"
)

var multiGenerator = generator.NewMulti()

func Generate(r reconciliation.Reconciliation) error {
	if err := multiGenerator.Generate(r, "Certificate"); err != nil {
		return err
	}
	return nil
}
