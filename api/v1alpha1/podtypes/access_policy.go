package podtypes

// AccessPolicy
//
// Zero trust dictates that only applications with a reason for being able
// to access another resource should be able to reach it. This is set up by
// default by denying all ingress and egress traffic from the Pods in the
// Deployment. The AccessPolicy field is an allowlist of other applications and hostnames
// that are allowed to talk with this Application and which resources this app can talk to
//
// +kubebuilder:object:generate=true
type AccessPolicy struct {
	// Inbound specifies the ingress rules. Which apps on the cluster can talk to this app?
	//
	//+kubebuilder:validation:Optional
	Inbound *InboundPolicy `json:"inbound,omitempty"`

	// Outbound specifies egress rules. Which apps on the cluster and the
	// internet is the Application allowed to send requests to?
	//
	//+kubebuilder:validation:Optional
	Outbound OutboundPolicy `json:"outbound,omitempty"`
}

// InboundPolicy
//
// +kubebuilder:object:generate=true
type InboundPolicy struct {
	// The rules list specifies a list of applications. When no namespace is
	// specified it refers to an app in the current namespace. For apps in
	// other namespaces namespace is required
	//
	//+kubebuilder:validation:Required
	Rules []InternalRule `json:"rules"`
}

// OutboundPolicy
//
// The rules list specifies a list of applications that are reachable on the cluster.
// Note that the application you're trying to reach also must specify that they accept communication
// from this app in their ingress rules.
//
// +kubebuilder:object:generate=true
type OutboundPolicy struct {
	// Rules apply the same in-cluster rules as InboundPolicy
	//
	//+kubebuilder:validation:Optional
	Rules []InternalRule `json:"rules,omitempty"`

	// External specifies which applications on the internet the application
	// can reach. Only host is required unless it is on another port than HTTPS port 443.
	// If other ports or protocols are required then `ports` must be specified as well
	//
	//+kubebuilder:validation:Optional
	External []ExternalRule `json:"external,omitempty"`
}

// InternalRule
//
// The rules list specifies a list of applications. When no namespace is
// specified it refers to an app in the current namespace. For apps in
// other namespaces namespace is required.
// If you add both Namespace and NamespacesByLabel to an InternalRule,
// Namespace takes presedence and NamespacesByLabel is omitted.
//
// +kubebuilder:validation:Optional
type InternalRule struct {
	// The namespace in which the Application you are allowing traffic to/from resides. If unset, uses namespace of Application.
	//
	//+kubebuilder:validation:Optional
	Namespace string `json:"namespace,omitempty"`

	// The name of the Application you are allowing traffic to/from.
	//
	//+kubebuilder:validation:Required
	Application string `json:"application"`
	//+kubebuilder:validation:Optional
	Namespace string `json:"namespace,omitempty"`
	//+kubebuilder:validation:Optional
	NamespacesByLabel map[string]string `json:"namespacesByLabel,omitempty"`
}

// ExternalRule
//
// Describes a rule for allowing your Application to route traffic to external applications and hosts.
//
// +kubebuilder:object:generate=true
type ExternalRule struct {
	// The allowed hostname. Note that this does not include subdomains.
	//
	//+kubebuilder:validation:Required

	Host string `json:"host"`
	// Non-HTTP requests (i.e. using the TCP protocol) need to use IP in addition to hostname
	// Only required for TCP requests.
	//
	// Note: Hostname must always be defined even if IP is set statically
	//
	//+kubebuilder:validation:Optional
	Ip string `json:"ip,omitempty"`

	// The ports to allow for the above hostname. When not specified HTTP and
	// HTTPS on port 80 and 443 respectively are put into the allowlist
	//
	//+kubebuilder:validation:Optional
	Ports []ExternalPort `json:"ports,omitempty"`
}

// ExternalPort
//
// A custom port describing an external host
type ExternalPort struct {
	// Name is required and is an arbitrary name. Must be unique within all ExternalRule ports.
	//
	//+kubebuilder:validation:Required
	Name string `json:"name"`

	// The port number of the external host
	//
	//+kubebuilder:validation:Required
	Port int `json:"port"`

	// The protocol to use for communication with the host. Only HTTP, HTTPS and TCP are supported.
	//
	//+kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=HTTP;HTTPS;TCP
	Protocol string `json:"protocol"`
}
