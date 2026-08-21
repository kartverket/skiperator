package gwapi

import (
	"context"
	"fmt"
	"slices"
	"strings"

	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gatewayapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

func validateApplicationConflicts(ctx context.Context, c client.Client, application *skiperatorv1alpha1.Application) error {
	if !application.UsesStandardRouting() {
		return nil
	}
	hosts, err := application.Hostnames()
	if err != nil {
		return err
	}

	listenerSets := &gatewayapiv1.ListenerSetList{}
	if err := c.List(ctx, listenerSets); err != nil {
		return fmt.Errorf("failed to list Gateway API ListenerSets: %w", err)
	}
	for _, host := range hosts.AllHosts() {
		for _, listenerSet := range listenerSets.Items {
			if !skiperatorManaged(listenerSet.Labels) || sameApplication(listenerSet.Labels, application) {
				continue
			}
			if listenerSet.Spec.ParentRef.Name != GatewayNameForHost(host.Hostname) {
				continue
			}
			for _, listener := range listenerSet.Spec.Listeners {
				if listenerCoversHostname(listener.Hostname, host.Hostname) {
					if listenerSetAccepted(listenerSet) {
						return fmt.Errorf("hostname %q already has an accepted ListenerSet %s/%s", host.Hostname, listenerSet.Namespace, listenerSet.Name)
					}
					return fmt.Errorf("hostname %q already has a pending ListenerSet %s/%s", host.Hostname, listenerSet.Namespace, listenerSet.Name)
				}
			}
		}
	}
	return nil
}

// RoutePathConflict is an overlap between one of a Routing's path prefixes and
// an accepted HTTPRoute belonging to another Routing on the same hostname.
type RoutePathConflict struct {
	Hostname   string
	PathPrefix string
	Route      types.NamespacedName
}

func (c RoutePathConflict) Message() string {
	return fmt.Sprintf("path %q on hostname %q overlaps accepted HTTPRoute %s, so Gateway API decides per request which of the two serves it", c.PathPrefix, c.Hostname, c.Route)
}

// validateRoutingConflicts enforces first-accepted-route-wins for shared
// hostnames.
//
// This is what makes path-based product-team routing under one hostname
// predictable. Skiperator refuses overlapping path prefixes only when the
// existing HTTPRoute is already accepted by Gateway API. Redirect-only routes
// are ignored because they do not claim a backend path.
//
// The exception is a path this Routing is already serving. Two Routings applied
// close together both pass validation, and once Gateway API accepts them the
// overlap exists whatever Skiperator does. Refusing the reconcile then hands the
// path to nobody and only stops this Routing from converging, so the overlap is
// reported through DetectRoutingPathConflict instead. Claiming a new path over
// an accepted route is still refused, including from a Routing that already
// serves other paths on the hostname.
func validateRoutingConflicts(ctx context.Context, c client.Client, routing *skiperatorv1alpha1.Routing) error {
	if !routing.UsesStandardRouting() {
		return nil
	}
	hosts, err := routing.Hostnames()
	if err != nil {
		return err
	}
	host := hosts.AllHosts()[0]
	if err := validateRoutingHostnameOwnership(ctx, c, routing, host.Hostname); err != nil {
		return err
	}
	conflict, alreadyServing, err := findRoutingPathConflict(ctx, c, routing, host.Hostname)
	if err != nil {
		return err
	}
	if conflict == nil || alreadyServing {
		return nil
	}
	return fmt.Errorf("path %q on hostname %q conflicts with accepted HTTPRoute %s", conflict.PathPrefix, conflict.Hostname, conflict.Route)
}

// DetectRoutingPathConflict reports the overlap that validateRoutingConflicts
// lets through. Gateway API has already picked which route answers a request,
// and it tells neither team, so this is what the contributors have to go on.
func DetectRoutingPathConflict(ctx context.Context, c client.Client, routing *skiperatorv1alpha1.Routing) (*RoutePathConflict, error) {
	if !routing.UsesStandardRouting() {
		return nil, nil
	}
	hosts, err := routing.Hostnames()
	if err != nil {
		return nil, err
	}
	conflict, _, err := findRoutingPathConflict(ctx, c, routing, hosts.AllHosts()[0].Hostname)
	return conflict, err
}

// findRoutingPathConflict returns the first overlap between routing's paths and
// another Routing's accepted route on hostname. The second return says whether
// routing's own accepted route already carries the conflicting path, which is
// what separates a live overlap from a new claim on someone else's path.
func findRoutingPathConflict(ctx context.Context, c client.Client, routing *skiperatorv1alpha1.Routing, hostname string) (*RoutePathConflict, bool, error) {
	routes := &gatewayapiv1.HTTPRouteList{}
	if err := c.List(ctx, routes); err != nil {
		return nil, false, fmt.Errorf("failed to list Gateway API HTTPRoutes: %w", err)
	}

	var conflict *RoutePathConflict
	servedPaths := []string{}
	for _, existing := range routes.Items {
		if !skiperatorManaged(existing.Labels) || isRedirectRoute(existing) || !routeHasHostname(existing, hostname) {
			continue
		}
		if sameRouting(existing.Labels, routing) {
			if routeAccepted(existing) {
				servedPaths = append(servedPaths, routePathPrefixes(existing)...)
			}
			continue
		}
		if conflict != nil || !routeAccepted(existing) {
			continue
		}
		if pathPrefix, overlaps := overlappingPathPrefix(existing, routing); overlaps {
			conflict = &RoutePathConflict{
				Hostname:   hostname,
				PathPrefix: pathPrefix,
				Route:      types.NamespacedName{Namespace: existing.Namespace, Name: existing.Name},
			}
		}
	}
	if conflict == nil {
		return nil, false, nil
	}
	return conflict, slices.Contains(servedPaths, conflict.PathPrefix), nil
}

// overlappingPathPrefix returns the first path prefix in routing's spec that
// collides with a rule on existing.
func overlappingPathPrefix(existing gatewayapiv1.HTTPRoute, routing *skiperatorv1alpha1.Routing) (string, bool) {
	for _, rule := range existing.Spec.Rules {
		for _, route := range routing.Spec.Routes {
			if routeRuleOverlaps(rule, route.PathPrefix) {
				return route.PathPrefix, true
			}
		}
	}
	return "", false
}

// routePathPrefixes lists the path prefixes a live HTTPRoute matches on.
func routePathPrefixes(route gatewayapiv1.HTTPRoute) []string {
	prefixes := []string{}
	for _, rule := range route.Spec.Rules {
		for _, match := range rule.Matches {
			if match.Path != nil && match.Path.Value != nil {
				prefixes = append(prefixes, *match.Path.Value)
			}
		}
	}
	return prefixes
}

// validateRoutingHostnameOwnership prevents standalone Routing from attaching
// to a hostname already claimed by shared or application-owned ListenerSets.
func validateRoutingHostnameOwnership(ctx context.Context, c client.Client, routing *skiperatorv1alpha1.Routing, hostname string) error {
	listenerSets := &gatewayapiv1.ListenerSetList{}
	if err := c.List(ctx, listenerSets); err != nil {
		return fmt.Errorf("failed to list Gateway API ListenerSets: %w", err)
	}
	for _, listenerSet := range listenerSets.Items {
		if !skiperatorManaged(listenerSet.Labels) || sameRouting(listenerSet.Labels, routing) {
			continue
		}
		if listenerSet.Spec.ParentRef.Name != GatewayNameForHost(hostname) {
			continue
		}
		for _, listener := range listenerSet.Spec.Listeners {
			if !listenerCoversHostname(listener.Hostname, hostname) {
				continue
			}
			if routing.UsesSharedOwnership() && sharedRoutingListenerSet(listenerSet.Labels, listenerSet.Name, hostname) {
				continue
			}
			if listenerSetAccepted(listenerSet) {
				return fmt.Errorf("hostname %q already has an accepted ListenerSet %s/%s", hostname, listenerSet.Namespace, listenerSet.Name)
			}
			return fmt.Errorf("hostname %q already has a pending ListenerSet %s/%s", hostname, listenerSet.Namespace, listenerSet.Name)
		}
	}
	return nil
}

func listenerSetAccepted(listenerSet gatewayapiv1.ListenerSet) bool {
	return meta.IsStatusConditionTrue(listenerSet.Status.Conditions, string(gatewayapiv1.ListenerSetConditionAccepted))
}

func routeAccepted(route gatewayapiv1.HTTPRoute) bool {
	for _, parent := range route.Status.Parents {
		if meta.IsStatusConditionTrue(parent.Conditions, string(gatewayapiv1.RouteConditionAccepted)) {
			return true
		}
	}
	return false
}

func skiperatorManaged(labels map[string]string) bool {
	return labels["app.kubernetes.io/managed-by"] == "skiperator"
}

func sameApplication(labels map[string]string, application *skiperatorv1alpha1.Application) bool {
	return labels["skiperator.kartverket.no/controller"] == "application" &&
		labels["application.skiperator.no/app-name"] == application.Name &&
		labels["application.skiperator.no/app-namespace"] == application.Namespace
}

func sameRouting(labels map[string]string, routing *skiperatorv1alpha1.Routing) bool {
	return labels["skiperator.kartverket.no/controller"] == "routing" &&
		labels["skiperator.kartverket.no/routing-name"] == routing.Name &&
		labels["skiperator.kartverket.no/source-namespace"] == routing.Namespace
}

// sharedRoutingListenerSet identifies the one shared ListenerSet that all
// shared Routing objects for a hostname are allowed to reuse.
func sharedRoutingListenerSet(labels map[string]string, name string, hostname string) bool {
	return labels["skiperator.kartverket.no/controller"] == "routing-shared" &&
		name == SharedListenerSetName(hostname)
}

func isRedirectRoute(route gatewayapiv1.HTTPRoute) bool {
	for _, rule := range route.Spec.Rules {
		if len(rule.BackendRefs) > 0 {
			return false
		}
		hasRedirect := false
		for _, filter := range rule.Filters {
			if filter.Type != gatewayapiv1.HTTPRouteFilterRequestRedirect {
				return false
			}
			hasRedirect = true
		}
		if !hasRedirect {
			return false
		}
	}
	return len(route.Spec.Rules) > 0
}

func routeHasHostname(route gatewayapiv1.HTTPRoute, hostname string) bool {
	hostname = strings.ToLower(hostname)
	for _, h := range route.Spec.Hostnames {
		if strings.ToLower(string(h)) == hostname {
			return true
		}
	}
	return len(route.Spec.Hostnames) == 0
}

func listenerCoversHostname(listenerHostname *gatewayapiv1.Hostname, hostname string) bool {
	return listenerHostname == nil || strings.EqualFold(string(*listenerHostname), hostname)
}

// routeRuleOverlaps treats Gateway API PathPrefix matches as prefix trees.
// "/api" conflicts with "/api/v1" because both can match the same request.
// "/api" does not conflict with "/apiv2" because prefixes match path elements.
// See sigs.k8s.io/gateway-api/apis/v1/httproute_types.go, PathMatchPathPrefix.
func routeRuleOverlaps(rule gatewayapiv1.HTTPRouteRule, pathPrefix string) bool {
	if len(rule.Matches) == 0 {
		return true
	}
	for _, match := range rule.Matches {
		if match.Path == nil || match.Path.Value == nil {
			return pathPrefix == "/"
		}
		existing := *match.Path.Value
		if pathPrefixesOverlap(existing, pathPrefix) {
			return true
		}
	}
	return false
}

func pathPrefixesOverlap(a string, b string) bool {
	return pathPrefixContains(a, b) || pathPrefixContains(b, a)
}

func pathPrefixContains(prefix string, path string) bool {
	if !strings.HasPrefix(path, prefix) {
		return false
	}
	return len(path) == len(prefix) || strings.HasSuffix(prefix, "/") || path[len(prefix)] == '/'
}
