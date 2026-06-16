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

// ListenerSetName returns generated ListenerSet name for hostname.
func ListenerSetName(prefix string, hostname string) string {
	return fmt.Sprintf("%s-listener-%x", prefix, util.GenerateHashFromName(hostname))
}

// RouteName returns HTTPS backend HTTPRoute name.
func RouteName(prefix string) string {
	return prefix
}

// RedirectRouteName returns HTTP-to-HTTPS redirect HTTPRoute name.
func RedirectRouteName(prefix string) string {
	return fmt.Sprintf("%s-redirect", prefix)
}
