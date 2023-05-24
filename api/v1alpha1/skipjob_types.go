/*
Copyright 2023.

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
	"github.com/kartverket/skiperator/api/v1alpha1/podtypes"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// SKIPJobStatus defines the observed state of SKIPJob
type SKIPJobStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// SKIPJob is the Schema for the skipjobs API
type SKIPJob struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SKIPJobSpec   `json:"spec,omitempty"`
	Status SKIPJobStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// SKIPJobList contains a list of SKIPJob
type SKIPJobList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SKIPJob `json:"items"`
}

// SKIPJobSpec defines the desired state of SKIPJob
// +kubebuilder:object:generate=true
type SKIPJobSpec struct {

	//+kubebuilder:validation:Required
	Job JobSettings `json:"job"`

	//+kubebuilder:validation:Required
	Container ContainerSettings `json:"container"`
}

type ContainerSettings struct {
	//+kubebuilder:validation:Required
	Image string `json:"image"`

	//+kubebuilder:validation:Required
	Port int `json:"port"`

	//+kubebuilder:validation:Enum=low;medium;high
	//+kubebuilder:default=medium
	Priority string `json:"priority,omitempty"`
	//+kubebuilder:validation:Optional
	Command []string `json:"command,omitempty"`

	//+kubebuilder:validation:Optional
	Resources podtypes.ResourceRequirements `json:"resources,omitempty"`

	//+kubebuilder:validation:Optional
	Env []corev1.EnvVar `json:"env,omitempty"`
	//+kubebuilder:validation:Optional
	EnvFrom []podtypes.EnvFrom `json:"envFrom,omitempty"`
	//+kubebuilder:validation:Optional
	FilesFrom []podtypes.FilesFrom `json:"filesFrom,omitempty"`

	//+kubebuilder:validation:Optional
	AdditionalPorts []podtypes.InternalPort `json:"additionalPorts,omitempty"`
	//+kubebuilder:validation:Optional
	Liveness *podtypes.Probe `json:"liveness,omitempty"`
	//+kubebuilder:validation:Optional
	Readiness *podtypes.Probe `json:"readiness,omitempty"`
	//+kubebuilder:validation:Optional
	Startup *podtypes.Probe `json:"startup,omitempty"`

	//+kubebuilder:validation:Optional
	AccessPolicy podtypes.AccessPolicy `json:"accessPolicy,omitempty"`

	//+kubebuilder:validation:Optional
	GCP *podtypes.GCP `json:"gcp,omitempty"`

	//+kubebuilder:validation:Optional
	Labels map[string]string `json:"labels,omitempty"`

	//+kubebuilder:validation:Optional
	ResourceLabels map[string]map[string]string `json:"resourceLabels,omitempty"`
}

type JobSettings struct {
}

type CronSettings struct {
	//+kubebuilder:validation:Required
	Schedule string `json:"schedule"`

	ConcurrencyPolicy ConcurrencyPolicy `json:"allowConcurrency,omitempty"`

	Suspend bool `json:"suspend,omitempty"`
}

// +kubebuilder:validation:Enum=Allow,Forbid,Replace
// +kubebuilder:default="Allow"
type ConcurrencyPolicy string

const (
	ALLOW   ConcurrencyPolicy = "Allow"
	FORBID  ConcurrencyPolicy = "Forbid"
	REPLACE ConcurrencyPolicy = "Replace"
)
