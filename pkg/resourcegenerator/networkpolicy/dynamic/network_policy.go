package dynamic

import (
	"fmt"
	"github.com/kartverket/skiperator/pkg/reconciliation"
)

const (
	GrafanaAgentName      	= "grafana-agent"
	GrafanaAgentNamespace 	= GrafanaAgentName
	AlloyAgentName			= "alloy"
	AlloyAgentNamespace 	= "grafana-alloy"
)

func Generate(r reconciliation.Reconciliation) error {
	ctxLog := r.GetLogger()

	//TODO refactor more so we can have more common functions
	if r.GetType() == reconciliation.ApplicationType || r.GetType() == reconciliation.JobType {
		return generateForCommon(r)
	} else if r.GetType() == reconciliation.RoutingType {
		return generateForRouting(r)
	} else {
		err := fmt.Errorf("unsupported type %s in network policy", r.GetType())
		ctxLog.Error(err, "Failed to generate network policy")
		return err
	}
}
