package v1alpha1

import (
	"golang.org/x/exp/constraints"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

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

//+kubebuilder:object:generate=true
type ApplicationSpec struct {
	//+kubebuilder:validation:Required
	Image   string   `json:"image"`
	Command []string `json:"command,omitempty"`

	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
	Replicas  Replicas                    `json:"replicas,omitempty"`
	Strategy  Strategy                    `json:"strategy,omitempty"`

	Env       []corev1.EnvVar `json:"env,omitempty"`
	EnvFrom   []EnvFrom       `json:"envFrom,omitempty"`
	FilesFrom []FilesFrom     `json:"filesFrom,omitempty"`

	//+kubebuilder:validation:Required
	Port      int    `json:"port"`
	Liveness  *Probe `json:"liveness,omitempty"`
	Readiness *Probe `json:"readiness,omitempty"`
	Startup   *Probe `json:"startup,omitempty"`

	Ingresses    []string     `json:"ingresses,omitempty"`
	AccessPolicy AccessPolicy `json:"accessPolicy,omitempty"`
}

type Replicas struct {
	//+kubebuilder:validation:Required
	Min uint `json:"min"`
	Max uint `json:"max,omitempty"`

	TargetCpuUtilization uint `json:"targetCpuUtilization,omitempty"`
}

type Strategy struct {
	// +kubebuilder:validation:Enum=RollingUpdate;Recreate
	// +kubebuilder:default=RollingUpdate
	Type string `json:"type"`
}

type EnvFrom struct {
	ConfigMap string `json:"configMap,omitempty"`
	Secret    string `json:"secret,omitempty"`
}

type FilesFrom struct {
	//+kubebuilder:validation:Required
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

	//+kubebuilder:validation:Required
	Port uint16 `json:"port"`
	//+kubebuilder:validation:Required
	Path string `json:"path"`
}

//+kubebuilder:object:generate=true
type AccessPolicy struct {
	Inbound  InboundPolicy  `json:"inbound,omitempty"`
	Outbound OutboundPolicy `json:"outbound,omitempty"`
}

//+kubebuilder:object:generate=true
type InboundPolicy struct {
	Rules []InternalRule `json:"rules"`
}

//+kubebuilder:object:generate=true
type OutboundPolicy struct {
	Rules    []InternalRule `json:"rules,omitempty"`
	External []ExternalRule `json:"external,omitempty"`
}

type InternalRule struct {
	Namespace string `json:"namespace"`
	//+kubebuilder:validation:Required
	Application string `json:"application"`
}

//+kubebuilder:object:generate=true
type ExternalRule struct {
	//+kubebuilder:validation:Required
	Host  string `json:"host"`
	Ports []Port `json:"ports,omitempty"`
}

type Port struct {
	//+kubebuilder:validation:Required
	Name string `json:"name"`
	//+kubebuilder:validation:Required
	Port int `json:"port"`
	//+kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=HTTP;HTTPS
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
