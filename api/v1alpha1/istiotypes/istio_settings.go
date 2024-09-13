// This package reimplements specific parts of the Telemetry struct from "istio.io/api/telemetry/v1" in order to be
// forward compatible if we choose to incorporate the struct directly in the future.

package istiotypes

// Tracing contains relevant settings for tracing in the telemetry configuration
// +kubebuilder:object:generate=true
type Tracing struct {
	// NB: RandomSamplingPercentage uses a wrapped type of *wrappers.DoubleValue in the original struct, but due
	// to incompatibalities with the kubebuilder code generator, we have chosen to use a simple int instead. This only allows
	// for whole numbers, but this is sufficient for our use case.

	// RandomSamplingPercentage is the percentage of requests that should be sampled for tracing, specified by a whole number between 0-100.
	// Setting RandomSamplingPercentage to 0 will disable tracing.
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=100
	// +kubebuilder:default:=10
	RandomSamplingPercentage int `json:"randomSamplingPercentage,omitempty"`
}

// Telemetry is a placeholder for all relevant telemetry types, and may be extended in the future to configure additional telemetry settings.
//
// +kubebuilder:object:generate=true
type Telemetry struct {
	// Tracing is a list of tracing configurations for the telemetry resource. Normally only one tracing configuration is needed.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:={{randomSamplingPercentage: 10}}
	Tracing []*Tracing `json:"tracing,omitempty"`
}

// IstioSettings contains configuration settings for istio resources. Currently only telemetry configuration is supported.
//
// +kubebuilder:object:generate=true
type IstioSettings struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:={tracing: {{randomSamplingPercentage: 10}}}
	Telemetry Telemetry `json:"telemetry,omitempty"`
}
