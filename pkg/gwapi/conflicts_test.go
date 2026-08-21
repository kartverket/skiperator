package gwapi

import (
	"context"
	"testing"

	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	gatewayapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

func TestRouteRuleOverlapsUsesPathElementBoundaries(t *testing.T) {
	for _, tt := range []struct {
		name      string
		existing  string
		candidate string
		want      bool
	}{
		{name: "same path", existing: "/api", candidate: "/api", want: true},
		{name: "existing parent", existing: "/api", candidate: "/api/v1", want: true},
		{name: "candidate parent", existing: "/api/v1", candidate: "/api", want: true},
		{name: "sibling prefix", existing: "/api", candidate: "/apiv2", want: false},
		{name: "sibling longer existing", existing: "/apiv2", candidate: "/api", want: false},
		{name: "existing trailing slash", existing: "/api/", candidate: "/api/v1", want: true},
		{name: "candidate trailing slash", existing: "/api/v1", candidate: "/api/", want: true},
		{name: "root existing", existing: "/", candidate: "/api", want: true},
		{name: "root candidate", existing: "/api", candidate: "/", want: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, routeRuleOverlaps(routeRule(tt.existing), tt.candidate))
		})
	}
}

func routeRule(path string) gatewayapiv1.HTTPRouteRule {
	pathType := gatewayapiv1.PathMatchPathPrefix
	return gatewayapiv1.HTTPRouteRule{
		Matches: []gatewayapiv1.HTTPRouteMatch{
			{
				Path: &gatewayapiv1.HTTPPathMatch{
					Type:  &pathType,
					Value: &path,
				},
			},
		},
	}
}

// Two Routings applied close together both pass validation, because neither
// route is accepted yet when the other is checked. Once Gateway API accepts
// them the overlap is real and live, so it has to be reported rather than
// turned into an error state that stops a serving Routing from converging.
func TestRoutingPathConflictReportedOnceOwnRouteIsAccepted(t *testing.T) {
	ctx := context.Background()
	routing := sharedRouting("team-b", "ns-b", "/api/v1")

	t.Run("own route not accepted yet: refuse the claim", func(t *testing.T) {
		c := conflictClient(t, acceptedRoute("team-a", "ns-a", "/api"))
		require.ErrorContains(t, validateRoutingConflicts(ctx, c, routing), `path "/api/v1" on hostname "shared.example.com" conflicts with accepted HTTPRoute ns-a/team-a`)
	})

	t.Run("already serving the path: report instead of blocking", func(t *testing.T) {
		c := conflictClient(t, acceptedRoute("team-a", "ns-a", "/api"), ownAcceptedRoute(routing, "/api/v1"))
		require.NoError(t, validateRoutingConflicts(ctx, c, routing))

		conflict, err := DetectRoutingPathConflict(ctx, c, routing)
		require.NoError(t, err)
		require.NotNil(t, conflict)
		assert.Equal(t, "/api/v1", conflict.PathPrefix)
		assert.Equal(t, "ns-a/team-a", conflict.Route.String())
	})

	t.Run("accepted on another path: a new claim is still refused", func(t *testing.T) {
		c := conflictClient(t, acceptedRoute("team-a", "ns-a", "/api"), ownAcceptedRoute(routing, "/team-b"))
		require.ErrorContains(t, validateRoutingConflicts(ctx, c, routing), `conflicts with accepted HTTPRoute ns-a/team-a`)
	})

	t.Run("no overlap: no conflict", func(t *testing.T) {
		c := conflictClient(t, acceptedRoute("team-a", "ns-a", "/other"), ownAcceptedRoute(routing, "/api/v1"))
		require.NoError(t, validateRoutingConflicts(ctx, c, routing))

		conflict, err := DetectRoutingPathConflict(ctx, c, routing)
		require.NoError(t, err)
		assert.Nil(t, conflict)
	})
}

func conflictClient(t *testing.T, routes ...*gatewayapiv1.HTTPRoute) client.Client {
	t.Helper()
	s := runtime.NewScheme()
	require.NoError(t, gatewayapiv1.Install(s))
	builder := fake.NewClientBuilder().WithScheme(s)
	for _, route := range routes {
		builder = builder.WithObjects(route)
	}
	return builder.Build()
}

func sharedRouting(name string, namespace string, pathPrefix string) *skiperatorv1alpha1.Routing {
	return &skiperatorv1alpha1.Routing{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
		Spec: skiperatorv1alpha1.RoutingSpec{
			Hostname:        "shared.example.com",
			RoutingProvider: skiperatorv1alpha1.RoutingProviderStandard,
			Ownership:       skiperatorv1alpha1.RoutingOwnershipShared,
			Routes:          []skiperatorv1alpha1.Route{{PathPrefix: pathPrefix, TargetApp: name}},
		},
	}
}

func acceptedRoute(name string, namespace string, pathPrefix string) *gatewayapiv1.HTTPRoute {
	return &gatewayapiv1.HTTPRoute{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"app.kubernetes.io/managed-by":              "skiperator",
				"skiperator.kartverket.no/controller":       "routing",
				"skiperator.kartverket.no/routing-name":     name,
				"skiperator.kartverket.no/source-namespace": namespace,
			},
		},
		Spec: gatewayapiv1.HTTPRouteSpec{
			Hostnames: []gatewayapiv1.Hostname{"shared.example.com"},
			Rules:     []gatewayapiv1.HTTPRouteRule{backendRule(routeRule(pathPrefix))},
		},
		Status: gatewayapiv1.HTTPRouteStatus{
			RouteStatus: gatewayapiv1.RouteStatus{
				Parents: []gatewayapiv1.RouteParentStatus{{
					Conditions: []metav1.Condition{{
						Type:   string(gatewayapiv1.RouteConditionAccepted),
						Status: metav1.ConditionTrue,
						Reason: "Accepted",
					}},
				}},
			},
		},
	}
}

func ownAcceptedRoute(routing *skiperatorv1alpha1.Routing, pathPrefix string) *gatewayapiv1.HTTPRoute {
	return acceptedRoute(routing.Name, routing.Namespace, pathPrefix)
}

// backendRule turns a match-only rule into one that claims a backend, so it is
// not skipped as a redirect-only route.
func backendRule(rule gatewayapiv1.HTTPRouteRule) gatewayapiv1.HTTPRouteRule {
	rule.BackendRefs = []gatewayapiv1.HTTPBackendRef{{}}
	return rule
}
