package gwapi

import (
	"context"
	"fmt"
	"strings"

	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/api/meta"
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

// validateRoutingConflicts enforces first-accepted-route-wins for shared
// hostnames.
//
// This is what makes path-based product-team routing under one hostname
// predictable. Skiperator refuses overlapping path prefixes only when the
// existing HTTPRoute is already accepted by Gateway API. Redirect-only routes
// are ignored because they do not claim a backend path.
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
	routes := &gatewayapiv1.HTTPRouteList{}
	if err := c.List(ctx, routes); err != nil {
		return fmt.Errorf("failed to list Gateway API HTTPRoutes: %w", err)
	}
	for _, existing := range routes.Items {
		if !skiperatorManaged(existing.Labels) || sameRouting(existing.Labels, routing) || !routeAccepted(existing) || isRedirectRoute(existing) {
			continue
		}
		if !routeHasHostname(existing, host.Hostname) {
			continue
		}
		for _, existingRule := range existing.Spec.Rules {
			for _, newRoute := range routing.Spec.Routes {
				if routeRuleOverlaps(existingRule, newRoute.PathPrefix) {
					return fmt.Errorf("path %q on hostname %q conflicts with accepted HTTPRoute %s/%s", newRoute.PathPrefix, host.Hostname, existing.Namespace, existing.Name)
				}
			}
		}
	}
	return nil
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
