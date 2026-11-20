package gwapi

import (
	"fmt"

	"github.com/kartverket/skiperator/api/common"
	"github.com/kartverket/skiperator/pkg/mesh"
	"github.com/kartverket/skiperator/pkg/util"
	gatewayapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

const (
	// IstioGatewayNamespace is where the shared Gateway API resources live.
	IstioGatewayNamespace = mesh.GatewayNamespace

	InternalGatewayName = "istio-internal"
	ExternalGatewayName = "istio-external"
)

// GatewayNameForHost selects shared Kubernetes Gateway API Gateway by hostname.
// Deprecated: prefer GatewayNameForHostObj when a *common.Host is available so
// that ForceInternal is respected.
func GatewayNameForHost(hostname string) gatewayapiv1.ObjectName {
	if util.IsInternal(hostname) {
		return InternalGatewayName
	}
	return ExternalGatewayName
}

// GatewayNameForHostObj selects the shared Gateway API Gateway for a host,
// respecting the ForceInternal override in addition to the domain regex.
func GatewayNameForHostObj(h *common.Host) gatewayapiv1.ObjectName {
	if h.IsInternal() {
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

// CertificateReferenceGrantName returns the ReferenceGrant that lets one
// ListenerSet read a custom certificate in istio-gateways. Every namespace
// shares istio-gateways, so this name must be unique across all of them.
//
// The hash covers both parts, because joining them is not unique on its own. A
// namespace and a ListenerSet name are both DNS labels that can contain "-", so
// namespace "team-a" with name "app-x" and namespace "team" with name "a-app-x"
// join to one string. Two ListenerSets would then share a grant, and the second
// reconcile would point it at its own Secret. The first listener loses the
// authorization it needs to read its certificate. A slash cannot appear in
// either part, so it separates them in the hashed value.
func CertificateReferenceGrantName(listenerSetNamespace string, listenerSetName string) string {
	return fmt.Sprintf("%s-cert-%x", listenerSetNamespace, util.GenerateHashFromName(listenerSetNamespace+"/"+listenerSetName))
}

// RedirectRouteName returns HTTP-to-HTTPS redirect HTTPRoute name.
func RedirectRouteName(prefix string) string {
	return fmt.Sprintf("%s-redirect", prefix)
}

// SharedRedirectRouteName returns generated shared redirect HTTPRoute name for hostname.
func SharedRedirectRouteName(hostname string) string {
	return RedirectRouteName(fmt.Sprintf("shared-%x", util.GenerateHashFromName(hostname)))
}
