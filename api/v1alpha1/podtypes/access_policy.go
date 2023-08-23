package podtypes

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:generate=true
type AccessPolicy struct {
	//+kubebuilder:validation:Optional
	Inbound *InboundPolicy `json:"inbound,omitempty"`
	//+kubebuilder:validation:Optional
	Outbound OutboundPolicy `json:"outbound,omitempty"`
}

// +kubebuilder:object:generate=true
type InboundPolicy struct {
	//+kubebuilder:validation:Required
	Rules []InternalRule `json:"rules"`
}

// +kubebuilder:object:generate=true
type OutboundPolicy struct {
	//+kubebuilder:validation:Optional
	Rules []InternalRule `json:"rules,omitempty"`
	//+kubebuilder:validation:Optional
	External []ExternalRule `json:"external,omitempty"`
}

type InternalRule struct {
	//+kubebuilder:validation:Optional
	Namespace string `json:"namespace,omitempty"`
	//+kubebuilder:validation:Required
	Application string `json:"application"`
	//+kubebuilder:validation:Optional
	NamespacesByLabel *metav1.LabelSelector `json:"namespaceByLabel,omitempty"`
}

// +kubebuilder:object:generate=true
type ExternalRule struct {
	//+kubebuilder:validation:Required
	Host string `json:"host"`
	//+kubebuilder:validation:Optional
	Ip string `json:"ip,omitempty"`
	//+kubebuilder:validation:Optional
	Ports []ExternalPort `json:"ports,omitempty"`
}

type ExternalPort struct {
	//+kubebuilder:validation:Required
	Name string `json:"name"`
	//+kubebuilder:validation:Required
	Port int `json:"port"`
	//+kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=HTTP;HTTPS;TCP
	Protocol string `json:"protocol"`
}
