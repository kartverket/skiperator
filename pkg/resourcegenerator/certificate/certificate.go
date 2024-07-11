package certificate

import (
	"fmt"
	"github.com/kartverket/skiperator/pkg/reconciliation"
)

const (
	IstioGatewayNamespace = "istio-gateways"
	GrafanaAgentName      = "grafana-agent"
	GrafanaAgentNamespace = GrafanaAgentName
)

func Generate(r reconciliation.Reconciliation) error {
	ctxLog := r.GetLogger()

	//TODO refactor more so we can have more common functions
	if r.GetType() == reconciliation.ApplicationType {
		return generateForApplication(r)
	} else if r.GetType() == reconciliation.RoutingType {
		return generateForRouting(r)
	} else {
		err := fmt.Errorf("unsupported type %s in certificate", r.GetType())
		ctxLog.Error(err, "Failed to generate certificate")
		return err
	}
}
