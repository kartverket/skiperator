package podtypes

import (
	corev1 "k8s.io/api/core/v1"
)

// +kubebuilder:object:generate=true
type ResourceRequirements struct {
	// TODO
	// Remember to reasess whether or not Claims work properly with kubebuilder when we upgrade to Kubernetes 1.26

	//+kubebuilder:validation:Optional
	Limits corev1.ResourceList `json:"limits,omitempty"`

	//+kubebuilder:validation:Optional
	Requests corev1.ResourceList `json:"requests,omitempty"`
}
