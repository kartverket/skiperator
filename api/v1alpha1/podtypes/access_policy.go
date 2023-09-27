package podtypes

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

// If you add both Namespace and NamespacesByLabel to an InternalRule,
// Namespace takes presedence and NamespacesByLabel is omitted.
// +kubebuilder:object:generate=true
type InternalRule struct {
	//+kubebuilder:validation:Required
	Application string `json:"application"`
	//+kubebuilder:validation:Optional
	Namespace string `json:"namespace,omitempty"`
	//+kubebuilder:validation:Optional
	NamespacesByLabel map[string]string `json:"namespacesByLabel,omitempty"`
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
