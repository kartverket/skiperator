package istiotypes

// This package reimplements specific parts of the Telemetry struct from "istio.io/api/telemetry/v1" in order to be
// forward compatible if we choose to incorporate the struct directly in the future.

// Tracing contains relevant settings for tracing in the telemetry configuration
// NB: RandomSamplingPercentage actually uses a wrapped type of *wrappers.DoubleValue in the original struct, but due
// to incompatibalities with the kubebuilder code generator, we have chosen to use a simple int instead. This only allows
// for whole numbers, but this is sufficient for our use case.
//
// +kubebuilder:object:generate=true
type Tracing struct {
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=100
	// +kubebuilder:default:=10
	RandomSamplingPercentage int `json:"randomSamplingPercentage,omitempty"`
}

// Telemetry is a placeholder for all relevant telemetry types
//
// +kubebuilder:object:generate=true
type Telemetry struct {
	// +kubebuilder:validation:Optional
	Tracing []*Tracing `json:"tracing,omitempty"`
}

// IstioSettings contains configuration settings for istio resources
//
// +kubebuilder:object:generate=true
type IstioSettings struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:={tracing: {{randomSamplingPercentage: 10}}}
	Telemetry Telemetry `json:"telemetry"`
}
