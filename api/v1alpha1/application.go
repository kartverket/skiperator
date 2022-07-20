package v1alpha1

import (
	"golang.org/x/exp/constraints"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func init() {
	SchemeBuilder.Register(
		&ApplicationList{},
		&Application{},
	)
}

//+kubebuilder:object:root=true
type ApplicationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Application `json:"items"`
}

//+kubebuilder:object:root=true
//+kubebuilder:resource:shortName="app"
type Application struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ApplicationSpec `json:"spec,omitempty"`
}

type ApplicationSpec struct {
	//+kubebuilder:validation:Required
	Image   string   `json:"image"`
	Command []string `json:"command,omitempty"`

	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
	Replicas  Replicas                    `json:"replicas,omitempty"`

	Env       []corev1.EnvVar `json:"env,omitempty"`
	EnvFrom   []EnvFrom       `json:"envFrom,omitempty"`
	FilesFrom []FilesFrom     `json:"filesFrom,omitempty"`

	//+kubebuilder:validation:Required
	Port      int    `json:"port"`
	Liveness  *Probe `json:"liveness,omitempty"`
	Readiness *Probe `json:"readiness,omitempty"`

	Ingresses    []string     `json:"ingresses,omitempty"`
	AccessPolicy AccessPolicy `json:"accessPolicy,omitempty"`
}

type Replicas struct {
	Min uint `json:"min"`
	Max uint `json:"max,omitempty"`

	TargetCpuUtilization uint `json:"targetCpuUtilization,omitempty"`
}

type EnvFrom struct {
	ConfigMap string `json:"configMap,omitempty"`
	Secret    string `json:"secret,omitempty"`
}

type FilesFrom struct {
	MountPath string `json:"mountPath"`

	ConfigMap             string `json:"configMap,omitempty"`
	Secret                string `json:"secret,omitempty"`
	EmptyDir              string `json:"emptyDir,omitempty"`
	PersistentVolumeClaim string `json:"persistentVolumeClaim,omitempty"`
}

type Probe struct {
	InitialDelay     uint `json:"initialDelay,omitempty"`
	Timeout          uint `json:"timeout,omitempty"`
	FailureThreshold uint `json:"failureThreshold,omitempty"`

	Port uint16 `json:"port"`
	Path string `json:"path"`
}

type AccessPolicy struct {
	Inbound  InboundPolicy  `json:"inbound,omitempty"`
	Outbound OutboundPolicy `json:"outbound,omitempty"`
}

type InboundPolicy struct {
	Rules []InternalRule `json:"rules"`
}

type OutboundPolicy struct {
	Rules    []InternalRule `json:"rules"`
	External []ExternalRule `json:"external,omitempty"`
}

type InternalRule struct {
	//+kubebuilder:validation:Optional
	Namespace   string `json:"namespace"`
	Application string `json:"application"`
}

type ExternalRule struct {
	Host  string `json:"host"`
	Ports []Port `json:"ports,omitempty"`
}

type Port struct {
	Name     string `json:"name"`
	Port     int    `json:"port,omitempty"`
	Protocol string `json:"protocol"`
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
