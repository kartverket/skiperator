package v1alpha1

import (
	"golang.org/x/exp/constraints"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
type ApplicationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Application `json:"items"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName="app"
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.application.status`
type Application struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ApplicationSpec `json:"spec,omitempty"`

	Status ApplicationStatus `json:"status,omitempty"`
}

// +kubebuilder:object:generate=true
type ApplicationSpec struct {
	//+kubebuilder:validation:Required
	Image string `json:"image"`
	//+kubebuilder:validation:Optional
	Command []string `json:"command,omitempty"`

	//+kubebuilder:validation:Optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
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

// +kubebuilder:object:generate=true
type ApplicationStatus struct {
	TotalApplicationStatus       Status            `json:"application,omitempty"`
	ControllersApplicationStatus map[string]Status `json:"controllers,omitempty"`
}

// +kubebuilder:object:generate=true
type Status struct {
	Status    StatusNames `json:"status"`
	Message   string      `json:"message"`
	TimeStamp string      `json:"timestamp"`
}

type StatusNames string

const (
	SYNCED      StatusNames = "Synced"
	PROGRESSING StatusNames = "Progressing"
	ERROR       StatusNames = "Error"
	PENDING     StatusNames = "Pending"
)

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
