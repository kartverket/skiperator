package gwapi

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
