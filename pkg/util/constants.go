package util

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const SkiperatorUser = int64(150)

var (
	IstioMetricsPortNumber = intstr.FromInt32(15020)
	IstioMetricsPortName   = intstr.FromString("istio-metrics")
	IstioMetricsPath       = "/stats/prometheus"
	// IstioTraceProvider Name of the trace provider set up in the istiod installation
	IstioTraceProvider = "otel-tracing"

	IstioRevisionLabel = "istio.io/rev"

	DefaultMetricDropList = []string{
		"istio_request_bytes_bucket",
		"istio_response_bytes_bucket",
		"istio_request_duration_milliseconds_bucket",
	}
)

// A security context for use in pod containers created by Skiperator
// follow least privilege best practices for the whole Container Security Context
var LeastPrivilegeContainerSecurityContext = corev1.SecurityContext{
	Capabilities: PointTo(corev1.Capabilities{
		Add: []corev1.Capability{},
		Drop: []corev1.Capability{
			"ALL",
		},
	}),
	Privileged:               PointTo(false),
	RunAsUser:                PointTo(SkiperatorUser),
	RunAsGroup:               PointTo(SkiperatorUser),
	RunAsNonRoot:             PointTo(true),
	ReadOnlyRootFilesystem:   PointTo(true),
	AllowPrivilegeEscalation: PointTo(false),
	SeccompProfile: &corev1.SeccompProfile{
		Type: corev1.SeccompProfileTypeRuntimeDefault,
	},
}
