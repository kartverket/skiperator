package v1beta1

import (
	"fmt"
	"time"

	"github.com/nais/liberator/pkg/namegen"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation"
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
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.summary.status`
// +kubebuilder:storageversion
type Routing struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	//+kubebuilder:validation:Required
	Spec   RoutingSpec      `json:"spec,omitempty"`
	Status SkiperatorStatus `json:"status,omitempty"`
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
	//+kubebuilder:validation:Optional
	Port int32 `json:"port,omitempty"`
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
	// https://github.com/nais/naiserator/blob/faed273b68dff8541e1e2889fda5d017730f9796/pkg/resourcecreator/idporten/idporten.go#L82
	// https://github.com/nais/naiserator/blob/faed273b68dff8541e1e2889fda5d017730f9796/pkg/resourcecreator/idporten/idporten.go#L170
	secretName, err := namegen.ShortName(fmt.Sprintf("%s-%s", namePrefix, "routing-ingress"), validation.DNS1035LabelMaxLength)
	return secretName, err
}

func (in *Routing) GetConditions() []metav1.Condition {
	return in.Status.Conditions
}

func (in *Routing) SetConditions(conditions []metav1.Condition) {
	in.Status.Conditions = conditions
}

func (in *RoutingSpec) GetHost() (*Host, error) {
	return NewHost(in.Hostname)
}

func (in *Routing) GetStatus() *SkiperatorStatus {
	return &in.Status
}

func (in *Routing) SetStatus(status SkiperatorStatus) {
	in.Status = status
}

func (in *Routing) GetDefaultLabels() map[string]string {
	return map[string]string{
		"app.kubernetes.io/name":                    in.Name,
		"app.kubernetes.io/managed-by":              "skiperator",
		"skiperator.kartverket.no/controller":       "routing",
		"skiperator.kartverket.no/routing-name":     in.Name,
		"skiperator.kartverket.no/source-namespace": in.Namespace,
	}
}

func (in *Routing) GetCommonSpec() *CommonSpec {
	panic("common spec not available for routing resource type")
}

func (in *Routing) SetDefaultStatus() {
	var msg string

	if in.Status.Summary.Status == "" {
		msg = "Default Routing status, it has not initialized yet"
	} else {
		msg = "Routing is trying to reconcile"
	}

	in.Status.Summary = Status{
		Status:    PENDING,
		Message:   msg,
		TimeStamp: time.Now().String(),
	}

	if in.Status.SubResources == nil {
		in.Status.SubResources = make(map[string]Status)
	}

	if len(in.Status.Conditions) == 0 {
		in.Status.Conditions = make([]metav1.Condition, 0)
	}
}
