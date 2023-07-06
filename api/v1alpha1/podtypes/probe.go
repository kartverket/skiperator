package podtypes

type Probe struct {
	//+kubebuilder:default=0
	//+kubebuilder:validation:Optional
	InitialDelay int32 `json:"initialDelay,omitempty"`
	//+kubebuilder:default=1
	//+kubebuilder:validation:Optional
	Timeout int32 `json:"timeout,omitempty"`
	//+kubebuilder:default=10
	//+kubebuilder:validation:Optional
	Period int32 `json:"period,omitempty"`
	//+kubebuilder:default=1
	//+kubebuilder:validation:Optional
	SuccessThreshold int32 `json:"successThreshold,omitempty"`
	//+kubebuilder:default=3
	//+kubebuilder:validation:Optional
	FailureThreshold int32 `json:"failureThreshold,omitempty"`
	//+kubebuilder:validation:Required
	Port uint16 `json:"port"`
	//+kubebuilder:validation:Required
	Path string `json:"path"`
}
