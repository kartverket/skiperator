package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

//+kubebuilder:object:root=true

// RoutingList contains a list of Routing
type RoutingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Routing `json:"items"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName="routing"
type Routing struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	//+kubebuilder:validation:Required
	Spec RoutingSpec `json:"spec,omitempty"`

	//+kubebuilder:validation:Optional
	Status RoutingStatus `json:"status,omitempty"`
}

// +kubebuilder:object:generate=true
type RoutingSpec struct {
	//+kubebuilder:validation:Required
	Hostname string `json:"hostname"`

	//+kubebuilder:validation:Required
	Routes []Route `json:"routes"`

	//+kubebuilder:validation:Optional
	//+kubebuilder:default:=true
	RedirectToHTTPS *bool `json:"redirectToHTTPS,omitempty"`
}

// +kubebuilder:object:generate=true
type Route struct {
	//+kubebuilder:validation:Required
	TargetApp string `json:"targetApp"`
	//+kubebuilder:validation:Required
	PathPrefix string `json:"pathPrefix"`
	//+kubebuilder:validation:Optional
	//+kubebuilder:default:=false
	RewriteUri bool `json:"rewriteUri,omitempty"`
}

// +kubebuilder:object:generate=true
type RoutingStatus struct {
	//+kubebuilder:validation:Optional
	Conditions []metav1.Condition `json:"conditions"`
}

// Get RedirectToHTTPS
func (in *Routing) GetRedirectToHTTPS() bool {
	if in.Spec.RedirectToHTTPS != nil {
		return *in.Spec.RedirectToHTTPS
	}
	return true
}
