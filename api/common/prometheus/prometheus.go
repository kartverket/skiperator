package common

import "k8s.io/apimachinery/pkg/util/intstr"

// PrometheusConfig contains configuration settings instructing how the app should be scraped.
//
// +kubebuilder:object:generate=true
type PrometheusConfig struct {
	// The port number or name where metrics are exposed (at the Pod level).
	//
	//+kubebuilder:validation:Required
	Port intstr.IntOrString `json:"port"`
	// The HTTP path where Prometheus compatible metrics exists
	//
	//+kubebuilder:default:=/metrics
	//+kubebuilder:validation:Optional
	Path string `json:"path,omitempty"`

	// Setting AllowAllMetrics to true will ensure all exposed metrics are scraped. Otherwise, a list of predefined
	// metrics will be dropped by default. See util/constants.go for the default list.
	//
	//+kubebuilder:default:=false
	//+kubebuilder:validation:Optional
	AllowAllMetrics bool `json:"allowAllMetrics,omitempty"`

	// ScrapeInterval specifies the interval at which Prometheus should scrape the metrics.
	// The interval must be at least 15 seconds (if using "Xs") and divisible by 5.
	// If minutes ("Xm") are used, the value must be at least 1m.
	//
	//+kubebuilder:default:="60s"
	//+kubebuilder:validation:Optional
	//+kubebuilder:validation:XValidation:rule="self == '' || self.matches('^([0-9]+[sm])+$')",messageExpression="'Rejected: ' + self + ' as an invalid value. ScrapeInterval must be empty (default applies) or in the format of <number>s or <number>m.'"
	//+kubebuilder:validation:XValidation:rule="self == '' || (self.endsWith('m') && int(self.split('m')[0]) >= 1) || (self.endsWith('s') && int(self.split('s')[0]) >= 15 && int(self.split('s')[0]) % 5 == 0)",messageExpression="'Rejected: ' + self + ' as an invalid value. ScrapeInterval must be at least 15s (if using <s>) and divisible by 5, or at least 1m (if using <m>).'"
	ScrapeInterval string `json:"scrapeInterval,omitempty"`
}
