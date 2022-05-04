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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type NetworkPolicy struct {
	// Setting this to true allows traffic from the ingress gateway of the
	// cluster to the application
	AcceptIngressTraffic bool `json:"acceptIngressTraffic"`
}

// SkipSpec defines the desired state of Skip
type SkipSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	NetworkPolicies []NetworkPolicy `json:"networkPolicies"`
}

// SkipStatus defines the observed state of Skip
type SkipStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Skip is the Schema for the skips API
type Skip struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SkipSpec   `json:"spec,omitempty"`
	Status SkipStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// SkipList contains a list of Skip
type SkipList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Skip `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Skip{}, &SkipList{})
}
