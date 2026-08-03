package gwapi

import (
	"fmt"

	"github.com/kartverket/skiperator/pkg/util"
	gatewayapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

const (
	IstioGatewayNamespace = "istio-gateways"
	InternalGatewayName   = "istio-internal"
	ExternalGatewayName   = "istio-external"
)

// GatewayNameForHost selects shared Kubernetes Gateway API Gateway by hostname.
func GatewayNameForHost(hostname string) gatewayapiv1.ObjectName {
	if util.IsInternal(hostname) {
		return InternalGatewayName
	}
	return ExternalGatewayName
}

// RoutingResourcePrefix qualifies a Routing's name for its Gateway API
// resources, so a standalone Routing and an Application with the same name in
// one namespace cannot collide on HTTPRoute/ListenerSet names. This mirrors the
// legacy "<name>-routing-ingress" naming convention. Applications keep the bare
// name as the primary case.
func RoutingResourcePrefix(name string) string {
	return fmt.Sprintf("%s-routing", name)
}

// ListenerSetName returns generated ListenerSet name for hostname.
func ListenerSetName(prefix string, hostname string) string {
	return fmt.Sprintf("%s-listener-%x", prefix, util.GenerateHashFromName(hostname))
}

// SharedListenerSetName returns generated shared ListenerSet name for hostname.
func SharedListenerSetName(hostname string) string {
	return ListenerSetName("shared", hostname)
}

// RedirectRouteName returns HTTP-to-HTTPS redirect HTTPRoute name.
func RedirectRouteName(prefix string) string {
	return fmt.Sprintf("%s-redirect", prefix)
}

// SharedRedirectRouteName returns generated shared redirect HTTPRoute name for hostname.
func SharedRedirectRouteName(hostname string) string {
	return RedirectRouteName(fmt.Sprintf("shared-%x", util.GenerateHashFromName(hostname)))
}
