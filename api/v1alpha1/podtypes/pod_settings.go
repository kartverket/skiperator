package podtypes

// PodSettings
//
// +kubebuilder:object:generate=true
type PodSettings struct {
	// Annotations that are set on Pods created by Skiperator. These annotations can for example be used to change the behaviour of sidecars and similar.
	//
	//+kubebuilder:validation:Optional
	Annotations map[string]string `json:"annotations,omitempty"`

	// TerminationGracePeriodSeconds determines how long Kubernetes waits after a SIGTERM signal sent to a Pod before terminating the pod. If your application uses longer than
	// 30 seconds to terminate, you should increase TerminationGracePeriodSeconds.
	//
	//+kubebuilder:validation:Optional
	//+kubebuilder:default:=30
	TerminationGracePeriodSeconds int64 `json:"terminationGracePeriodSeconds,omitempty"`
}
