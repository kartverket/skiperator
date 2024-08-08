package certificate

import (
	"github.com/kartverket/skiperator/pkg/reconciliation"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/resourceutils/generator"
)

const (
	IstioGatewayNamespace = "istio-gateways"
	GrafanaAgentName      = "grafana-agent"
	GrafanaAgentNamespace = GrafanaAgentName
)

var multiGenerator = generator.NewMulti()

func Generate(r reconciliation.Reconciliation) error {
	return multiGenerator.Generate(r, "Certificate")
}
