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
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SKIPJobStatus defines the observed state of SKIPJob
// +kubebuilder:object:generate=true
type SKIPJobStatus struct {
	Conditions []metav1.Condition `json:"conditions"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:object:generate=true
// SKIPJob is the Schema for the skipjobs API
type SKIPJob struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	//+kubebuilder:validation:Required
	Spec   SKIPJobSpec   `json:"spec"`
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
	// Settings for the actual Job. If you use a scheduled job, the settings in here will also specify the template of the job.
	//
	//+kubebuilder:validation:Optional
	Job *JobSettings `json:"job,omitempty"`

	// Settings for the Job if you are running a scheduled job. Optional as Jobs may be one-off.
	//
	//+kubebuilder:validation:Optional
	Cron *CronSettings `json:"cron,omitempty"`

	// Settings for the Pods running in the job. Fields are mostly the same as an Application, and are (probably) better documented there. Some fields are omitted, but none added.
	//
	//+kubebuilder:validation:Required
	Container ContainerSettings `json:"container"`
}

// +kubebuilder:object:generate=true
type ContainerSettings struct {
	//+kubebuilder:validation:Required
	Image string `json:"image"`

	//+kubebuilder:validation:Enum=low;medium;high
	//+kubebuilder:default=medium
	Priority string `json:"priority,omitempty"`
	//+kubebuilder:validation:Optional
	Command []string `json:"command,omitempty"`

	//+kubebuilder:validation:Optional
	Resources *podtypes.ResourceRequirements `json:"resources,omitempty"`

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
	AccessPolicy *podtypes.AccessPolicy `json:"accessPolicy,omitempty"`

	//+kubebuilder:validation:Optional
	GCP *podtypes.GCP `json:"gcp,omitempty"`

	//+kubebuilder:validation:Optional
	Labels map[string]string `json:"labels,omitempty"`

	//+kubebuilder:validation:Optional
	ResourceLabels map[string]map[string]string `json:"resourceLabels,omitempty"`

	// +kubebuilder:validation:Enum=OnFailure;Never
	// +kubebuilder:default="Never"
	// +kubebuilder:validation:Optional
	RestartPolicy *corev1.RestartPolicy `json:"restartPolicy"`
}

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

	// Specifies the number of Pods to run in parallel. Defaults to 1. Works similar to a Skiperator Application's Replicas.
	//
	//+kubebuilder:validation:Optional
	Parallelism *int32 `json:"parallelism,omitempty"`

	// If set to true, this tells Kubernetes to suspend this Job till the field is set to false. If the Job is active while this field is set to false,
	// all running Pods will be terminated.
	//
	//+kubebuilder:validation:Optional
	//+kubebuilder:default=false
	Suspend *bool `json:"suspend,omitempty"`

	// The number of seconds to wait before removing the Job after it has finished. If unset, Job will not be cleaned up.
	// It is recommended to set this to avoid clutter in your resource tree.
	//
	//+kubebuilder:validation:Optional
	TTLSecondsAfterFinished *int32 `json:"ttlSecondsAfterFinished,omitempty"`

	// Settings for managing the Hook lifecycle of the Job, if wanted.
	//
	//+kubebuilder:validation:Optional
	HookSettings *HookSettings `json:"hookSettings,omitempty"`
}

// HookSettings
// Created own settings object to allow for configuration of more hook settings in the future, just in case.
//
// +kubebuilder:object:generate=true
type HookSettings struct {
	// Sets the SyncPhase of the Job. Only PreSync and PostSync are allowed. If unset, a normal Sync is performed.
	// A PreSync will wait for the Job to complete before syncing other resources in the namespace, which might be useful for
	// things like database migrations. A PostSync will only be applied after all other resources are applied.
	//
	// +kubebuilder:validation:Enum=PreSync;PostSync
	//+kubebuilder:validation:Required
	SyncPhase *string `json:"syncPhase,omitempty"`

	// Potential configuration for DeletionPolicy
	//
	// DeletionPolicy *string `json:"deletionPolicy,omitempty"`
}

// +kubebuilder:object:generate=true
type CronSettings struct {
	// Denotes how Kubernetes should react to multiple instances of the Job being started at the same time.
	// Allow will allow concurrent jobs. Forbid will not allow this, and instead skip the newer schedule Job.
	// Replace will replace the current active Job with the newer scheduled Job.
	//
	// +kubebuilder:validation:Enum=Allow;Forbid;Replace
	// +kubebuilder:default="Allow"
	// +kubebuilder:validation:Optional
	ConcurrencyPolicy batchv1.ConcurrencyPolicy `json:"allowConcurrency,omitempty"`

	// A CronJob string for denoting the schedule of this job. See https://crontab.guru/ for help creating CronJob strings.
	// Kubernetes CronJobs also include the extended "Vixie cron" step values: https://man.freebsd.org/cgi/man.cgi?crontab%285%29.
	//
	//+kubebuilder:validation:Required
	Schedule string `json:"schedule"`

	// Denotes the deadline in seconds for starting a job on its schedule, if for some reason the Job's controller was not ready upon the scheduled time.
	// If unset, Jobs missing their deadline will be considered failed jobs and will not start.
	//
	//+kubebuilder:validation:Optional
	StartingDeadlineSeconds *int64 `json:"startingDeadlineSeconds,omitempty"`

	// If set to true, this tells Kubernetes to suspend this Job till the field is set to false. If the Job is active while this field is set to false,
	// all running Pods will be terminated.
	//
	//+kubebuilder:validation:Optional
	Suspend *bool `json:"suspend,omitempty"`
}
