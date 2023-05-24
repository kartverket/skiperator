package podtypes

type Probe struct {
	//+kubebuilder:validation:Optional
	InitialDelay uint `json:"initialDelay,omitempty"`
	//+kubebuilder:validation:Optional
	Timeout uint `json:"timeout,omitempty"`
	//+kubebuilder:validation:Optional
	FailureThreshold uint `json:"failureThreshold,omitempty"`

	//+kubebuilder:validation:Required
	Port uint16 `json:"port"`
	//+kubebuilder:validation:Required
	Path string `json:"path"`
}
