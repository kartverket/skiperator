/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// List of port rules for external communication. Must be specified if using protocols other than HTTPS.
type Port struct {
	// Human-readable identifier for this rule.
	Name string `json:"name"`
	// The port used for communication.
	Port int `json:"port,omitempty"`
	// The protocol used for communication. Allowed values: GRPC, HTTP, HTTP2, HTTPS, MONGO, TCP, TLS
	Protocol string `json:"protocol"`
}

type ExternalRule struct {
	Host  string `json:"host"`
	Ports []Port `json:"ports,omitempty"`
}

type Rule struct {
	Application string `json:"application"`
	Namespace   string `json:"namespace,omitempty"`
}

type InboundPolicy struct {
	Rules []Rule `json:"rules"`
}

type OutboundPolicy struct {
	Rules    []Rule         `json:"rules,omitempty"`
	External []ExternalRule `json:"external,omitempty"`
}

type AccessPolicy struct {
	Inbound  *InboundPolicy  `json:"inbound,omitempty"`
	Outbound *OutboundPolicy `json:"outbound,omitempty"`
}

type Replicas struct {
	CpuThresholdPercentage *int32 `json:"cpuThresholdPercentage,omitempty"`
	DisableAutoScaling     bool   `json:"disableAutoScaling,omitempty"`
	Max                    int32  `json:"max,omitempty"`
	Min                    *int32 `json:"min"`
}

type CpuMemory struct {
	Cpu    string `json:"cpu,omitempty"`
	Memory string `json:"memory,omitempty"`
}

type Resources struct {
	Limits   *CpuMemory `json:"limits,omitempty"`
	Requests *CpuMemory `json:"requests,omitempty"`
}

type Env struct {
	Name      string           `json:"name"`
	Value     string           `json:"value,omitempty"`
	ValueFrom *v1.EnvVarSource `json:"valueFrom,omitempty"`
}

type EnvFrom struct {
	Configmap string `json:"configmap,omitempty"`
	Secret    string `json:"secret,omitempty"`
}

type FilesFrom struct {
	MountPath             string `json:"mountPath"`
	Configmap             string `json:"configmap,omitempty"`
	PersistentVolumeClaim string `json:"persistentVolumeClaim,omitempty"`
	EmptyDir              string `json:"emptyDir,omitempty"`
	Secret                string `json:"secret,omitempty"`
}

type Probe struct {
	FailureThreshold int    `json:"failureThreshold,omitempty"`
	InitialDelay     int    `json:"initialDelay,omitempty"`
	Path             string `json:"path"`
	PeriodSeconds    int    `json:"periodSeconds,omitempty"`
	Port             int    `json:"port"`
	Timeout          int    `json:"timeout,omitempty"`
}

// ApplicationSpec defines the desired state of Application
type ApplicationSpec struct {
	// Important: Run "make" to regenerate code after modifying this file

	AccessPolicy *AccessPolicy `json:"accessPolicy,omitempty"`
	Image        string        `json:"image"`
	Port         int           `json:"port,omitempty"`
	Ingresses    []string      `json:"ingresses,omitempty"`
	Replicas     *Replicas     `json:"replicas,omitempty"`
	Resources    *Resources    `json:"resources,omitempty"`
	Env          []Env         `json:"env,omitempty"`
	EnvFrom      []EnvFrom     `json:"envFrom,omitempty"`
	FilesFrom    []FilesFrom   `json:"filesFrom,omitempty"`
	Liveness     *Probe        `json:"liveness,omitempty"`
	Readiness    *Probe        `json:"readiness,omitempty"`
}

// ApplicationStatus defines the observed state of Application
type ApplicationStatus struct {
	// Important: Run "make" to regenerate code after modifying this file
	OperationResults map[string]controllerutil.OperationResult `json:"operationResults"`
	Errors           map[string]string                         `json:"errors"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:shortName="app"

// Application is the Schema for the application API
type Application struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ApplicationSpec   `json:"spec,omitempty"`
	Status ApplicationStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ApplicationList contains a list of Application
type ApplicationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Application `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Application{}, &ApplicationList{})
}
