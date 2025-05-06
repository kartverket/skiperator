// This package reimplements specific parts of the Telemetry struct from "istio.io/api/telemetry/v1" in order to be
// forward compatible if we choose to incorporate the struct directly in the future.

package istiotypes

import (
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

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

// Retries is configurable automatic retries for requests towards the application.
// By default requests falling under: "connect-failure,refused-stream,unavailable,cancelled,5xx" will be retried.
//
// +kubebuilder:object:generate=true
type Retries struct {
	// Attempts is the number of retries to be allowed for a given request before giving up. The interval between retries will be determined automatically (25ms+).
	// Default is 2
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Optional
	Attempts *int32 `json:"attempts,omitempty"`

	// PerTryTimeout is the timeout per attempt for a given request, including the initial call and any retries. Format: 1h/1m/1s/1ms. MUST be >=1ms.
	// Default: no timeout
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Format=duration
	// +kubebuilder:validation:Optional
	PerTryTimeout *v1.Duration `json:"perTryTimeout,omitempty"`

	// RetryOnHttpResponseCodes HTTP response codes that should trigger a retry. A typical value is [503].
	// You may also use 5xx and retriable-4xx (only 409).
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:items:Enum="500";500;"501";501;"502";502;"503";503;"504";504;"505";505;"506";506;"507";507;"508";508;"510";510;"511";511;"409";409;"retriable-4xx";"5xx"
	RetryOnHttpResponseCodes *[]intstr.IntOrString `json:"retryOnHttpResponseCodes,omitempty"`
}

// IstioSettings contains configuration settings for istio resources. Currently only telemetry configuration is supported.
//
// +kubebuilder:object:generate=true
type IstioSettingsBase struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:={tracing: {{randomSamplingPercentage: 10}}}
	Telemetry Telemetry `json:"telemetry,omitempty"`
}

type IstioSettingsApplication struct {
	IstioSettingsBase `json:",inline"`

	// +kubebuilder:validation:Optional
	Retries *Retries `json:"retries,omitempty"`
}
