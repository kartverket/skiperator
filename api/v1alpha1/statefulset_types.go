package v1alpha1

import corev1 "k8s.io/api/core/v1"

// StatefulSpec configures the Application to be deployed as a StatefulSet
// instead of a Deployment. Requires VolumeClaimTemplates. Disallows
// Strategy.Type=Recreate and HPA-range replicas.
//
// +kubebuilder:object:generate=true
type StatefulSpec struct {
	// When true, generates a StatefulSet instead of a Deployment.
	// Requires VolumeClaimTemplates. Disallows Strategy.Type=Recreate and HPA-range replicas.
	// This value is immutable - delete and recreate the Application to change
	//
	//+kubebuilder:validation:Optional
	//+kubebuilder:default=false
	Enabled bool `json:"enabled,omitempty"`

	// Per-pod PersistentVolumeClaims provisioned by the StatefulSet controller.
	// Each replica gets its own PVC named `<template.metadata.name>-<app>-<ordinal>`.
	// Only valid when enabled=true
	//
	//+kubebuilder:validation:Optional
	VolumeClaimTemplates []VolumeClaimTemplate `json:"volumeClaimTemplates,omitempty"`

	// Controls pod creation and update order. OrderedReady creates pods one at a time, Parallel creates them simultaneously.
	// Only valid when enabled=true
	//
	//+kubebuilder:validation:Optional
	//+kubebuilder:validation:Enum=OrderedReady;Parallel
	PodManagementPolicy string `json:"podManagementPolicy,omitempty"`

	// Staged rollouts - only pods with ordinal >= Partition are updated.
	// Set Partition equal to replicas to pause updates.
	// Only valid when enabled=true
	//
	//+kubebuilder:validation:Optional
	//+kubebuilder:validation:Minimum=0
	Partition *int32 `json:"partition,omitempty"`

	// PVC fate when the StatefulSet is deleted. Defaults to Retain.
	// Only valid when enabled=true
	//
	//+kubebuilder:validation:Optional
	//+kubebuilder:validation:Enum=Retain;Delete
	PVCRetentionWhenDeleted string `json:"pvcRetentionWhenDeleted,omitempty"`

	// PVC fate when the StatefulSet is scaled down. Defaults to Retain.
	// Only valid when enabled=true
	//
	//+kubebuilder:validation:Optional
	//+kubebuilder:validation:Enum=Retain;Delete
	PVCRetentionWhenScaled string `json:"pvcRetentionWhenScaled,omitempty"`
}

// VolumeClaimTemplate describes a per-pod PersistentVolumeClaim provisioned by the StatefulSet
// controller. Name serves as both the pod volume reference and the PVC prefix
//
// +kubebuilder:object:generate=true
type VolumeClaimTemplate struct {
	// Pod volume name and PVC name prefix. Resulting PVCs are named `<name>-<app>-<ordinal>`
	//
	//+kubebuilder:validation:Required
	Name string `json:"name"`

	// PVC spec
	//
	//+kubebuilder:validation:Required
	Spec corev1.PersistentVolumeClaimSpec `json:"spec"`

	// Optional labels applied to PVCs
	//
	//+kubebuilder:validation:Optional
	Labels map[string]string `json:"labels,omitempty"`

	// Optional annotations applied to PVCs
	//
	//+kubebuilder:validation:Optional
	Annotations map[string]string `json:"annotations,omitempty"`

	// Where the volume is mounted inside the container
	//
	//+kubebuilder:validation:Required
	MountPath string `json:"mountPath"`

	// Subpath within the volume to mount instead of its root
	//
	//+kubebuilder:validation:Optional
	SubPath string `json:"subPath,omitempty"`
}
