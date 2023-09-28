package podtypes

import "k8s.io/apimachinery/pkg/util/intstr"

// Probe
//
// Type configuration for all types of Kubernetes probes.
type Probe struct {
	// Number of the port to access on the container
	//
	//+kubebuilder:validation:Required
	Port intstr.IntOrString `json:"port"`

	// The path to access on the HTTP server
	//
	//+kubebuilder:validation:Required
	Path string `json:"path"`

	// Delay sending the first probe by X seconds. Can be useful for applications that
	// are slow to start.
	//
	//+kubebuilder:default=0
	//+kubebuilder:validation:Optional
	InitialDelay int32 `json:"initialDelay,omitempty"`

	// Number of seconds after which the probe times out. Defaults to 1 second.
	// Minimum value is 1
	//
	//+kubebuilder:default=1
	//+kubebuilder:validation:Optional
	Timeout int32 `json:"timeout,omitempty"`

	// Number of seconds Kubernetes waits between each probe. Defaults to 10 seconds.
	//
	//+kubebuilder:default=10
	//+kubebuilder:validation:Optional
	Period int32 `json:"period,omitempty"`

	// Minimum consecutive successes for the probe to be considered successful after having failed.
	// Defaults to 1. Must be 1 for liveness and startup Probes. Minimum value is 1.
	//
	//+kubebuilder:default=1
	//+kubebuilder:validation:Optional
	SuccessThreshold int32 `json:"successThreshold,omitempty"`

	// Minimum consecutive failures for the probe to be considered failed after
	// having succeeded. Defaults to 3. Minimum value is 1
	//
	//+kubebuilder:default=3
	//+kubebuilder:validation:Optional
	FailureThreshold int32 `json:"failureThreshold,omitempty"`
}
