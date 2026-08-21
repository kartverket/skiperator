// Package mesh describes how a namespace joins the Istio mesh, and holds the
// labels, ports, and addresses that follow from that choice.
package mesh

import "k8s.io/apimachinery/pkg/util/intstr"

const (
	// RevisionLabel enables sidecar injection for a namespace.
	RevisionLabel = "istio.io/rev"

	// DataplaneModeLabel with the value AmbientDataplaneMode puts a namespace
	// in the ambient mesh. Ambient namespaces get no sidecar.
	DataplaneModeLabel   = "istio.io/dataplane-mode"
	AmbientDataplaneMode = "ambient"

	// GatewayNamespace holds the shared ingress gateways, and SystemNamespace
	// holds the control plane.
	GatewayNamespace = "istio-gateways"
	SystemNamespace  = "istio-system"

	// MetricsPath is where the sidecar serves its Prometheus metrics.
	MetricsPath = "/stats/prometheus"

	// TraceProvider is the name of the trace provider set up in the istiod
	// installation.
	TraceProvider = "otel-tracing"

	// AmbientHealthProbeCIDR is the link-local address ambient SNATs kubelet
	// health probes to, so that probe packets can be told apart from other
	// node traffic. IPv6 clusters also need
	// fd16:9254:7127:1337:ffff:ffff:ffff:ffff/128.
	// https://istio.io/latest/docs/ambient/usage/networkpolicy/
	AmbientHealthProbeCIDR = "169.254.7.127/32"
)

var (
	// ZtunnelInboundPort is the HBONE port ztunnel listens on. In ambient mode
	// inbound mesh traffic arrives here instead of on the application port.
	ZtunnelInboundPort = intstr.FromInt32(15008)

	// MetricsPortNumber and MetricsPortName address the sidecar metrics port.
	MetricsPortNumber = intstr.FromInt32(15020)
	MetricsPortName   = intstr.FromString("istio-metrics")

	// DefaultMetricDropList holds the high-cardinality sidecar histograms that
	// are dropped before ingestion.
	DefaultMetricDropList = []string{
		"istio_request_bytes_bucket",
		"istio_response_bytes_bucket",
		"istio_request_duration_milliseconds_bucket",
	}
)

// Mode tells how a namespace joins the mesh. The modes are exclusive: Istio
// does not enroll a pod that has a sidecar into the ambient mesh, so a
// namespace with both labels behaves as a sidecar namespace.
type Mode string

const (
	ModeNone    Mode = ""
	ModeSidecar Mode = "sidecar"
	ModeAmbient Mode = "ambient"
)

// ModeFromLabels reads the mode from the namespace labels.
func ModeFromLabels(labels map[string]string) Mode {
	if labels[RevisionLabel] != "" {
		return ModeSidecar
	}
	if labels[DataplaneModeLabel] == AmbientDataplaneMode {
		return ModeAmbient
	}
	return ModeNone
}

// IsEnabled reports whether Istio manages the namespace at all. Gateway API
// routing needs this, while sidecar-specific resources need ModeSidecar.
func (m Mode) IsEnabled() bool {
	return m != ModeNone
}

// IngressGatewayLabels selects the pods of the ingress gateway that serves a
// hostname. Internal and external hostnames get separate gateway deployments.
func IngressGatewayLabels(isInternal bool) map[string]string {
	if isInternal {
		return map[string]string{"app": "istio-ingress-internal"}
	}
	return map[string]string{"app": "istio-ingress-external"}
}

// GatewayNamespaceLabels selects the namespace the ingress gateways run in.
func GatewayNamespaceLabels() map[string]string {
	return map[string]string{"kubernetes.io/metadata.name": GatewayNamespace}
}
