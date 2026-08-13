package util

import corev1 "k8s.io/api/core/v1"

const SkiperatorUser = int64(150)

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
