package v1alpha1

import (
	"fmt"
	"github.com/kartverket/skiperator/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

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
	Spec   RoutingSpec   `json:"spec,omitempty"`
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

func (in *Routing) GetGatewayName() string {
	return fmt.Sprintf("%s-routing-ingress", in.Name)
}

func (in *Routing) GetVirtualServiceName() string {
	return fmt.Sprintf("%s-routing-ingress", in.Name)
}

func (in *Routing) GetCertificateName() (string, error) {
	namePrefix := fmt.Sprintf("%s-%s", in.Namespace, in.Name)
	return util.GetSecretName(namePrefix, "routing-ingress")
}

func (in *Routing) GetConditions() []metav1.Condition {
	return in.Status.Conditions
}

func (in *Routing) SetConditions(conditions []metav1.Condition) {
	in.Status.Conditions = conditions
}
