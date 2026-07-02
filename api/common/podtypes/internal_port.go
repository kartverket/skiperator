package podtypes

import corev1 "k8s.io/api/core/v1"

// +kubebuilder:validation:XValidation:rule="self.name != 'main' && self.name != 'istio-metrics'",message="port names 'main' and 'istio-metrics' are reserved"
type InternalPort struct {
	//+kubebuilder:validation:Required
	Name string `json:"name"`
	//+kubebuilder:validation:Required
	Port int32 `json:"port"`
	//+kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=TCP;UDP;SCTP
	// +kubebuilder:default:TCP
	Protocol corev1.Protocol `json:"protocol"`
}
