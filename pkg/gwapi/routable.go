package gwapi

import (
	"context"

	"github.com/kartverket/skiperator/api/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Routable is the small shared contract Application and Routing expose to
// Gateway API support code.
type Routable interface {
	client.Object
	UsesStandardRouting() bool
	Hostnames() (common.HostCollection, error)
	GetCertificateName(*common.Host) (string, error)
}

type legacyRoutable interface {
	client.Object
	GetVirtualServiceName() string
	GetGatewayNames() ([]string, error)
}

// ValidateConflicts checks Gateway API ownership rules before resources are
// generated.
//
// Application conflicts are hostname based because an Application owns its
// ListenerSets directly. Routing conflicts are hostname and path-prefix based
// because multiple Routing objects may intentionally share one hostname. The
// per-kind rule lives behind the planner, so this stays a thin dispatch.
func ValidateConflicts(ctx context.Context, c client.Client, routable Routable) error {
	planner, err := plannerFor(routable)
	if err != nil {
		return err
	}
	return planner.validateConflicts(ctx, c)
}

// EvaluateRoutingState inspects generated resources and returns the migration
// state used by reconcilers and resource generators.
//
// The important output is GenerateLegacyRouting. When true, legacy Istio
// Gateway and VirtualService resources must stay in the desired resource set.
// When false, resource processing may prune them. Standard routing readiness is
// based on Gateway API and certificate status, not on Skiperator object sync.
func EvaluateRoutingState(ctx context.Context, c client.Client, routable Routable, status *common.SkiperatorStatus) (RoutingStateResult, error) {
	var migrationStartedAt *metav1.Time
	if status != nil {
		migrationStartedAt = status.MigrationStartedAt
	}

	// The planner holds the only Application-vs-Routing type switch; from here on
	// the migration state machine is type-agnostic.
	planner, err := plannerFor(routable)
	if err != nil {
		return RoutingStateResult{}, err
	}

	if !routable.UsesStandardRouting() {
		return determineRoutingState(false, Readiness{}, false, migrationStartedAt), nil
	}
	legacyExists, err := legacyRoutingExists(ctx, c, planner)
	if err != nil {
		return RoutingStateResult{}, err
	}
	return determineRoutingState(true, observeStandardRouting(ctx, c, planner), legacyExists, migrationStartedAt), nil
}
