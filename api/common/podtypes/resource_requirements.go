package podtypes

import (
	corev1 "k8s.io/api/core/v1"
)

// ResourceRequirements
//
// A simplified version of the Kubernetes native ResourceRequirement field, in which only Limits and Requests are present.
// For the units used for resources, see https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/#resource-units-in-kubernetes
//
// +kubebuilder:object:generate=true
type ResourceRequirements struct {

	// Limits set the maximum the app is allowed to use. Exceeding this limit will
	// make kubernetes kill the app and restart it.
	//
	// Limits can be set on the CPU and memory, but it is not recommended to put a limit on CPU, see: https://home.robusta.dev/blog/stop-using-cpu-limits
	//
	//+kubebuilder:validation:Optional
	Limits corev1.ResourceList `json:"limits,omitempty"`

	// Requests set the initial allocation that is done for the app and will
	// thus be available to the app on startup. More is allocated on demand
	// until the limit is reached.
	//
	// Requests can be set on the CPU and memory.
	//
	//+kubebuilder:validation:Optional
	Requests corev1.ResourceList `json:"requests,omitempty"`
}
