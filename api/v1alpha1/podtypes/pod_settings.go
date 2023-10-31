package podtypes

// PodSettings
//
// +kubebuilder:object:generate=true
type PodSettings struct {
	//
	//
	//+kubebuilder:validation:Optional
	Annotations map[string]string `json:"annotations,omitempty"`

	//
	//
	//+kubebuilder:validation:Optional
	//+kubebuilder:default:=30
	TerminationGracePeriodSeconds int64 `json:"terminationGracePeriodSeconds,omitempty"`
}
