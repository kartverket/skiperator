package gwapi

import (
	"context"
	"fmt"

	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// routablePlanner centralizes the only Application-vs-Routing type switch in
// this package. Every per-type difference — Gateway API resource names, legacy
// resource refs, and conflict scope — lives behind this interface, so the rest
// of gwapi stays type-agnostic. Adding a new routable kind means adding one
// planner and one case in plannerFor, touching nothing else.
type routablePlanner interface {
	// legacyRoutable supplies the legacy Istio resources to probe during
	// migration; satisfied by the embedded concrete object.
	legacyRoutable
	// readinessPlan resolves all standard-routing probe targets without any
	// cluster I/O (the pure Plan stage).
	readinessPlan() (readinessPlan, error)
	// validateConflicts enforces the kind's ownership rules before resources
	// are generated.
	validateConflicts(ctx context.Context, c client.Client) error
}

func plannerFor(routable Routable) (routablePlanner, error) {
	switch obj := routable.(type) {
	case *skiperatorv1alpha1.Application:
		return applicationPlanner{obj}, nil
	case *skiperatorv1alpha1.Routing:
		return routingPlanner{obj}, nil
	default:
		return nil, fmt.Errorf("unsupported Gateway API routable type %T", routable)
	}
}

type applicationPlanner struct {
	*skiperatorv1alpha1.Application
}

func (p applicationPlanner) readinessPlan() (readinessPlan, error) {
	hosts, err := p.Hostnames()
	if err != nil {
		return readinessPlan{}, err
	}
	return buildReadinessPlan(planInput{
		namespace:       p.Namespace,
		routeBaseName:   p.Name,
		redirectToHTTPS: p.Spec.RedirectToHTTPS != nil && *p.Spec.RedirectToHTTPS,
		hosts:           hosts,
		certificateName: p.GetCertificateName,
	})
}

func (p applicationPlanner) validateConflicts(ctx context.Context, c client.Client) error {
	return validateApplicationConflicts(ctx, c, p.Application)
}

type routingPlanner struct {
	*skiperatorv1alpha1.Routing
}

func (p routingPlanner) readinessPlan() (readinessPlan, error) {
	hosts, err := p.Hostnames()
	if err != nil {
		return readinessPlan{}, err
	}
	return buildReadinessPlan(planInput{
		namespace:       p.Namespace,
		routeBaseName:   RoutingResourcePrefix(p.Name),
		redirectToHTTPS: p.GetRedirectToHTTPS(),
		hosts:           hosts,
		certificateName: p.GetCertificateName,
		sharedRouting:   p.UsesSharedOwnership(),
	})
}

func (p routingPlanner) validateConflicts(ctx context.Context, c client.Client) error {
	return validateRoutingConflicts(ctx, c, p.Routing)
}
