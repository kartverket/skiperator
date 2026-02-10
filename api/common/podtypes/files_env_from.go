package podtypes

type EnvFrom struct {
	// Name of Kubernetes ConfigMap in which the deployment should mount environment variables from. Must be in the same namespace as the Application
	//
	//+kubebuilder:validation:Optional
	ConfigMap string `json:"configMap,omitempty"`

	// Name of Kubernetes Secret in which the deployment should mount environment variables from. Must be in the same namespace as the Application
	//
	//+kubebuilder:validation:Optional
	Secret string `json:"secret,omitempty"`
}

// FilesFrom
//
// Struct representing information needed to mount a Kubernetes resource as a file to a Pod's directory.
// One of ConfigMap, Secret, EmptyDir or PersistentVolumeClaim must be present, and just represent the name of the resource in question
// NB. Out-of-the-box, skiperator provides a writable 'emptyDir'-volume at '/tmp'
type FilesFrom struct {
	// The path to mount the file in the Pods directory. Required.
	//
	//+kubebuilder:validation:Required
	MountPath string `json:"mountPath"`

	//+kubebuilder:validation:Optional
	ConfigMap string `json:"configMap,omitempty"`
	//+kubebuilder:validation:Optional
	Secret string `json:"secret,omitempty"`
	//+kubebuilder:validation:Optional
	EmptyDir string `json:"emptyDir,omitempty"`
	//+kubebuilder:validation:Optional
	PersistentVolumeClaim string `json:"persistentVolumeClaim,omitempty"`
	// defaultMode is optional: mode bits used to set permissions on created files by default.
	// Must be an octal value between 0000 and 0777 or a decimal value between 0 and 511.
	// YAML accepts both octal and decimal values, JSON requires decimal values for mode bits.
	// Defaults to 0644.
	// Directories within the path are not affected by this setting.
	// This might be in conflict with other options that affect the file
	// mode, like fsGroup, and the result can be other mode bits set.
	// +kubebuilder:validation:min=0
	// +kubebuilder:validation:max=777
	// +kubebuilder:validation:Optional
	DefaultMode int `json:"defaultMode,omitempty"`
}
