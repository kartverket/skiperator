// +k8s:deepcopy-gen=package
package common

// +kubebuilder:object:generate=true
type JobSettings struct {
	// ActiveDeadlineSeconds denotes a duration in seconds started from when the job is first active. If the deadline is reached during the job's workload
	// the job and its Pods are terminated. If the job is suspended using the Suspend field, this timer is stopped and reset when unsuspended.
	//
	//+kubebuilder:validation:Optional
	ActiveDeadlineSeconds *int64 `json:"activeDeadlineSeconds,omitempty"`

	// Specifies the number of retry attempts before determining the job as failed. Defaults to 6.
	//
	//+kubebuilder:validation:Optional
	BackoffLimit *int32 `json:"backoffLimit,omitempty"`

	// If set to true, this tells Kubernetes to suspend this Job till the field is set to false. If the Job is active while this field is set to false,
	// all running Pods will be terminated.
	//
	//+kubebuilder:validation:Optional
	Suspend *bool `json:"suspend,omitempty"`

	// The number of seconds to wait before removing the Job after it has finished. If unset, Job will not be cleaned up.
	// It is recommended to set this to avoid clutter in your resource tree.
	//
	//+kubebuilder:validation:Optional
	TTLSecondsAfterFinished *int32 `json:"ttlSecondsAfterFinished,omitempty"`
}
