package gwapi

import (
	"context"
	"fmt"

	"github.com/kartverket/skiperator/api/common"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
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
// because multiple Routing objects may intentionally share one hostname.
func ValidateConflicts(ctx context.Context, c client.Client, routable Routable) error {
	switch obj := routable.(type) {
	case *skiperatorv1alpha1.Application:
		return validateApplicationConflicts(ctx, c, obj)
	case *skiperatorv1alpha1.Routing:
		return validateRoutingConflicts(ctx, c, obj)
	default:
		return fmt.Errorf("unsupported Gateway API routable type %T", routable)
	}
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

	switch obj := routable.(type) {
	case *skiperatorv1alpha1.Application:
		if !obj.UsesStandardRouting() {
			return determineRoutingState(false, Readiness{}, false, migrationStartedAt), nil
		}
		legacyExists, err := legacyRoutingExists(ctx, c, obj)
		if err != nil {
			return RoutingStateResult{}, err
		}
		return determineRoutingState(
			true,
			applicationStandardRoutingReady(ctx, c, obj),
			legacyExists,
			migrationStartedAt,
		), nil
	case *skiperatorv1alpha1.Routing:
		if !obj.UsesStandardRouting() {
			return determineRoutingState(false, Readiness{}, false, migrationStartedAt), nil
		}
		legacyExists, err := legacyRoutingExists(ctx, c, obj)
		if err != nil {
			return RoutingStateResult{}, err
		}
		return determineRoutingState(
			true,
			routingStandardRoutingReady(ctx, c, obj),
			legacyExists,
			migrationStartedAt,
		), nil
	default:
		return RoutingStateResult{}, fmt.Errorf("unsupported Gateway API routable type %T", routable)
	}
}
