package util

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var CommonAnnotations = map[string]string{
	// Prevents Argo CD from deleting these resources and leaving the namespace
	// in a deadlocked deleting state
	// https://argo-cd.readthedocs.io/en/stable/user-guide/sync-options/#no-prune-resources
	"argocd.argoproj.io/sync-options": "Prune=false",
}

const SkiperatorUser = int64(150)

var (
	IstioMetricsPortNumber = intstr.FromInt(15020)
	IstioMetricsPortName   = intstr.FromString("istio-metrics")
	IstioMetricsPath       = "/stats/prometheus"

	IstioRevisionLabel = "istio.io/rev"
)

// A security context for use in pod containers created by Skiperator
// follow least privilege best practices for the whole Container Security Context
var LeastPrivilegeContainerSecurityContext = corev1.SecurityContext{
	Capabilities: PointTo(corev1.Capabilities{
		Add: []corev1.Capability{},
		Drop: []corev1.Capability{
			"all",
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
