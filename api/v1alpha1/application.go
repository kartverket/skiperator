package v1alpha1

import (
	"golang.org/x/exp/constraints"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
type ApplicationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Application `json:"items"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:shortName="app"
type Application struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ApplicationSpec `json:"spec,omitempty"`
}

// +kubebuilder:object:generate=true
type ResourceRequirements struct {
	//+kubebuilder:validation:Required
	Limits *LimitResourceList `json:"limits"`
	//+kubebuilder:validation:Required
	Requests *RequestsResourceList `json:"requests"`
}

// +kubebuilder:object:generate=true
type LimitResourceList struct {
	//+kubebuilder:validation:Optional
	CpuLimit resource.Quantity `json:"cpu,omitempty"`
	//+kubebuilder:validation:Required
	MemoryLimit resource.Quantity `json:"memory"`
	//+kubebuilder:validation:Optional
	StorageLimit resource.Quantity `json:"storage,omitempty"`
	//+kubebuilder:validation:Optional
	EphemeralStorageLimit resource.Quantity `json:"ephemeralStorage,omitempty"`
}

// +kubebuilder:object:generate=true
type RequestsResourceList struct {
	//+kubebuilder:validation:Required
	CpuRequest resource.Quantity `json:"cpu"`
	//+kubebuilder:validation:Required
	MemoryRequest resource.Quantity `json:"memory"`
	//+kubebuilder:validation:Optional
	StorageRequest resource.Quantity `json:"storage,omitempty"`
	//+kubebuilder:validation:Optional
	EphemeralStorageRequest resource.Quantity `json:"ephemeralStorage,omitempty"`
}

// +kubebuilder:object:generate=true
type ApplicationSpec struct {
	//+kubebuilder:validation:Required
	Image string `json:"image"`
	//+kubebuilder:validation:Optional
	Command []string `json:"command,omitempty"`

	//+kubebuilder:validation:Optional
	Resources *ResourceRequirements `json:"resources,omitempty"`
	//+kubebuilder:validation:Optional
	Replicas Replicas `json:"replicas,omitempty"`
	//+kubebuilder:validation:Optional
	Strategy Strategy `json:"strategy,omitempty"`

	//+kubebuilder:validation:Optional
	Env []corev1.EnvVar `json:"env,omitempty"`
	//+kubebuilder:validation:Optional
	EnvFrom []EnvFrom `json:"envFrom,omitempty"`
	//+kubebuilder:validation:Optional
	FilesFrom []FilesFrom `json:"filesFrom,omitempty"`

	//+kubebuilder:validation:Required
	Port int `json:"port"`
	//+kubebuilder:validation:Optional
	Liveness *Probe `json:"liveness,omitempty"`
	//+kubebuilder:validation:Optional
	Readiness *Probe `json:"readiness,omitempty"`
	//+kubebuilder:validation:Optional
	Startup *Probe `json:"startup,omitempty"`

	//+kubebuilder:validation:Optional
	Ingresses []string `json:"ingresses,omitempty"`
	//+kubebuilder:validation:Optional
	AccessPolicy AccessPolicy `json:"accessPolicy,omitempty"`

	//+kubebuilder:validation:Optional
	GCP *GCP `json:"gcp,omitempty"`
}

type Replicas struct {
	//+kubebuilder:validation:Required
	Min uint `json:"min"`
	//+kubebuilder:validation:Optional
	Max uint `json:"max,omitempty"`

	//+kubebuilder:validation:Optional
	TargetCpuUtilization uint `json:"targetCpuUtilization,omitempty"`
}

type Strategy struct {
	//+kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum=RollingUpdate;Recreate
	// +kubebuilder:default=RollingUpdate
	Type string `json:"type"`
}

type EnvFrom struct {
	//+kubebuilder:validation:Optional
	ConfigMap string `json:"configMap,omitempty"`
	//+kubebuilder:validation:Optional
	Secret string `json:"secret,omitempty"`
}

type FilesFrom struct {
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
}

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

// +kubebuilder:object:generate=true
type AccessPolicy struct {
	//+kubebuilder:validation:Optional
	Inbound InboundPolicy `json:"inbound,omitempty"`
	//+kubebuilder:validation:Optional
	Outbound OutboundPolicy `json:"outbound,omitempty"`
}

// +kubebuilder:object:generate=true
type InboundPolicy struct {
	//+kubebuilder:validation:Optional
	Rules []InternalRule `json:"rules"`
}

// +kubebuilder:object:generate=true
type OutboundPolicy struct {
	//+kubebuilder:validation:Optional
	Rules []InternalRule `json:"rules,omitempty"`
	//+kubebuilder:validation:Optional
	External []ExternalRule `json:"external,omitempty"`
}

type InternalRule struct {
	//+kubebuilder:validation:Optional
	Namespace string `json:"namespace"`
	//+kubebuilder:validation:Required
	Application string `json:"application"`
}

// +kubebuilder:object:generate=true
type ExternalRule struct {
	//+kubebuilder:validation:Required
	Host string `json:"host"`
	//+kubebuilder:validation:Optional
	Ip string `json:"ip"`
	//+kubebuilder:validation:Optional
	Ports []Port `json:"ports,omitempty"`
}

type Port struct {
	//+kubebuilder:validation:Required
	Name string `json:"name"`
	//+kubebuilder:validation:Required
	Port int `json:"port"`
	//+kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=HTTP;HTTPS;TCP
	Protocol string `json:"protocol"`
}

type GCP struct {
	//+kubebuilder:validation:Required
	Auth Auth `json:"auth"`
}

type Auth struct {
	//+kubebuilder:validation:Required
	ServiceAccount string `json:"serviceAccount"`
}

func (a *Application) FillDefaults() {
	a.Spec.Replicas.Min = max(1, a.Spec.Replicas.Min)
	a.Spec.Replicas.Max = max(a.Spec.Replicas.Min, a.Spec.Replicas.Max)

	if a.Spec.Replicas.TargetCpuUtilization == 0 {
		a.Spec.Replicas.TargetCpuUtilization = 80
	}
}

func max[T constraints.Ordered](a, b T) T {
	if a > b {
		return a
	} else {
		return b
	}
}

func (a *Application) LimitToCoreResourceList(resList *LimitResourceList) corev1.ResourceList {
	resourceList := make(corev1.ResourceList)

	if !resList.CpuLimit.IsZero() {
		resourceList["cpu"] = resList.CpuLimit
	}

	resourceList["memory"] = resList.MemoryLimit

	if !resList.StorageLimit.IsZero() {
		resourceList["storage"] = resList.StorageLimit
	}

	if !resList.EphemeralStorageLimit.IsZero() {
		resourceList["ephemeral-storage"] = resList.EphemeralStorageLimit
	}

	return resourceList
}

func (a *Application) RequestToCoreResourceList(resList *RequestsResourceList) corev1.ResourceList {
	resourceList := make(corev1.ResourceList)

	resourceList["cpu"] = resList.CpuRequest
	resourceList["memory"] = resList.MemoryRequest

	if !resList.StorageRequest.IsZero() {
		resourceList["storage"] = resList.StorageRequest
	}

	if !resList.EphemeralStorageRequest.IsZero() {
		resourceList["ephemeral-storage"] = resList.EphemeralStorageRequest
	}

	return resourceList
}
