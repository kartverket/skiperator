package certificate

import (
	"github.com/kartverket/skiperator/v3/pkg/reconciliation"
	"github.com/kartverket/skiperator/v3/pkg/resourcegenerator/resourceutils/generator"
)

const (
	IstioGatewayNamespace = "istio-gateways"
)

var multiGenerator = generator.NewMulti()

func Generate(r reconciliation.Reconciliation) error {
	return multiGenerator.Generate(r, "Certificate")
}
