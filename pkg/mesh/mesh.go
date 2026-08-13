// Package mesh describes how a namespace joins the Istio mesh, and holds the
// labels, ports, and addresses that follow from that choice.
package mesh

import "k8s.io/apimachinery/pkg/util/intstr"

const (
	// RevisionLabel enables sidecar injection for a namespace.
	RevisionLabel = "istio.io/rev"

	// GatewayNamespace holds the shared ingress gateways, and SystemNamespace
	// holds the control plane.
	GatewayNamespace = "istio-gateways"
	SystemNamespace  = "istio-system"

	// MetricsPath is where the sidecar serves its Prometheus metrics.
	MetricsPath = "/stats/prometheus"

	// TraceProvider is the name of the trace provider set up in the istiod
	// installation.
	TraceProvider = "otel-tracing"
)

var (
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
