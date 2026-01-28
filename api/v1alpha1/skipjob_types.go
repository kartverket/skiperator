package v1alpha1

import (
	"fmt"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	DefaultTTLSecondsAfterFinished = int32(60 * 60 * 24 * 7) // One week
	DefaultBackoffLimit            = int32(6)

	DefaultSuspend           = false
	ConditionRunning         = "Running"
	ConditionFinished        = "Finished"
	ConditionFailed          = "Failed"
	SKIPJobReferenceLabelKey = "skiperator.kartverket.no/skipjobName"
	IsSKIPJobKey             = "skiperator.kartverket.no/skipjob"
)

// SKIPJobStatus defines the observed state of SKIPJob
// +kubebuilder:object:generate=true
type SKIPJobStatus struct {
	//+kubebuilder:validation:Optional
	Conditions []metav1.Condition `json:"conditions"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:object:generate=true
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.summary.status`
// +kubebuilder:printcolumn:name="AccessPolicies",type=string,JSONPath=`.status.accessPolicies`
//
// SKIPJob is the Schema for the skipjobs API
type SKIPJob struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	//+kubebuilder:validation:Required
	Spec SKIPJobSpec `json:"spec"`

	//+kubebuilder:validation:Optional
	Status SkiperatorStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// SKIPJobList contains a list of SKIPJob
type SKIPJobList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SKIPJob `json:"items"`
}

// SKIPJobSpec defines the desired state of SKIPJob
//
// +kubebuilder:object:generate=true
// A SKIPJob is either defined as a one-off or a scheduled job. If the Cron field is set for SKIPJob, it may not be removed. If the Cron field is unset, it may not be added.
// The Container field of a SKIPJob is only mutable if the Cron field is set. If unset, you must delete your SKIPJob to change container settings.
// +kubebuilder:validation:XValidation:rule="(has(oldSelf.cron) && has(self.cron)) || (!has(oldSelf.cron) && !has(self.cron))", message="After creation of a SKIPJob you may not remove the Cron field if it was previously present, or add it if it was previously omitted. Please delete the SKIPJob to change its nature from a one-off/scheduled job."
// +kubebuilder:validation:XValidation:rule="((!has(self.cron) && (oldSelf.container == self.container)) || has(self.cron))", message="The field Container is immutable for one-off jobs. Please delete your SKIPJob to change the containers settings."
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
	// Once set, you may not change Container without deleting your current SKIPJob
	//
	// +kubebuilder:validation:Required
	Container ContainerSettings `json:"container"`

	// IstioSettings are used to configure istio specific resources such as telemetry. Currently, adjusting sampling
	// interval for tracing is the only supported option.
	// By default, tracing is enabled with a random sampling percentage of 10%.
	//
	//+kubebuilder:validation:Optional
	//+kubebuilder:default:={telemetry: {tracing: {{randomSamplingPercentage: 10}}}}
	IstioSettings *IstioSettingsBase `json:"istioSettings,omitempty"`

	// Prometheus settings for pod running in job. Fields are identical to Application and if set,
	// a podmonitoring object is created.
	Prometheus *PrometheusConfig `json:"prometheus,omitempty"`
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
	Resources *ResourceRequirements `json:"resources,omitempty"`

	//+kubebuilder:validation:Optional
	Env []corev1.EnvVar `json:"env,omitempty"`
	//+kubebuilder:validation:Optional
	EnvFrom []EnvFrom `json:"envFrom,omitempty"`
	//+kubebuilder:validation:Optional
	FilesFrom []FilesFrom `json:"filesFrom,omitempty"`

	//+kubebuilder:validation:Optional
	AdditionalPorts []InternalPort `json:"additionalPorts,omitempty"`
	//+kubebuilder:validation:Optional
	Liveness *Probe `json:"liveness,omitempty"`
	//+kubebuilder:validation:Optional
	Readiness *Probe `json:"readiness,omitempty"`
	//+kubebuilder:validation:Optional
	Startup *Probe `json:"startup,omitempty"`

	//+kubebuilder:validation:Optional
	AccessPolicy *AccessPolicy `json:"accessPolicy,omitempty"`

	//+kubebuilder:validation:Optional
	GCP *GCP `json:"gcp,omitempty"`

	// +kubebuilder:validation:Enum=OnFailure;Never
	// +kubebuilder:default="Never"
	// +kubebuilder:validation:Optional
	RestartPolicy *corev1.RestartPolicy `json:"restartPolicy"`

	//+kubebuilder:validation:Optional
	PodSettings *PodSettings `json:"podSettings,omitempty"`
}

func (skipJob *SKIPJob) KindPostFixedName() string {
	return strings.ToLower(fmt.Sprintf("%v-%v", skipJob.Name, skipJob.Kind))
}

func (skipJob *SKIPJob) GetStatus() *SkiperatorStatus {
	return &skipJob.Status
}
func (skipJob *SKIPJob) SetStatus(status SkiperatorStatus) {
	skipJob.Status = status
}

func (skipJob *SKIPJob) FillDefaultSpec() {
	if skipJob.Spec.Job == nil {
		skipJob.Spec.Job = &JobSettings{}
	}

	if skipJob.Spec.Job.TTLSecondsAfterFinished == nil {
		skipJob.Spec.Job.TTLSecondsAfterFinished = &DefaultTTLSecondsAfterFinished
	}

	if skipJob.Spec.Job.BackoffLimit == nil {
		skipJob.Spec.Job.BackoffLimit = &DefaultBackoffLimit
	}

	if skipJob.Spec.Job.Suspend == nil {
		skipJob.Spec.Job.Suspend = &DefaultSuspend
	}

	if skipJob.Spec.Cron != nil {
		if skipJob.Spec.Cron.Suspend == nil {
			skipJob.Spec.Cron.Suspend = &DefaultSuspend
		}
	}
}

// TODO we should test SKIPJob status better, same for Routing probably
func (skipJob *SKIPJob) FillDefaultStatus() {
	var msg string

	if skipJob.Status.Summary.Status == "" {
		msg = "Default SKIPJob status, it has not initialized yet"
	} else {
		msg = "SKIPJob is trying to reconcile"
	}

	skipJob.Status.Summary = Status{
		Status:    PENDING,
		Message:   msg,
		TimeStamp: time.Now().String(),
	}

	if skipJob.Status.SubResources == nil {
		skipJob.Status.SubResources = make(map[string]Status)
	}

	if len(skipJob.Status.Conditions) == 0 {
		skipJob.Status.Conditions = []metav1.Condition{
			{
				Type:               ConditionRunning,
				Status:             metav1.ConditionFalse,
				LastTransitionTime: metav1.Now(),
				Reason:             "NotReconciled",
				Message:            "SKIPJob has not been reconciled yet",
			},
			{
				Type:               ConditionFinished,
				Status:             metav1.ConditionFalse,
				LastTransitionTime: metav1.Now(),
				Reason:             "NotReconciled",
				Message:            "SKIPJob has not been reconciled yet",
			},
			{
				Type:               ConditionFailed,
				Status:             metav1.ConditionFalse,
				LastTransitionTime: metav1.Now(),
				Reason:             "NotReconciled",
				Message:            "SKIPJob has not been reconciled yet",
			},
		}
	}
}

func (skipJob *SKIPJob) GetDefaultLabels() map[string]string {
	return map[string]string{
		"app.kubernetes.io/name":              skipJob.Name,
		"app.kubernetes.io/managed-by":        "skiperator",
		"skiperator.kartverket.no/controller": "skipjob",
		// Used by hahaha to know that the Pod should be watched for killing sidecars
		IsSKIPJobKey: "true",
		// Added to be able to add the SKIPJob to a reconcile queue when Watched Jobs are queued
		SKIPJobReferenceLabelKey: skipJob.Name,
	}
}

func (skipJob *SKIPJob) GetCommonSpec() *CommonSpec {
	return &CommonSpec{
		GCP:           skipJob.Spec.Container.GCP,
		AccessPolicy:  skipJob.Spec.Container.AccessPolicy,
		IstioSettings: skipJob.Spec.IstioSettings,
		Image:         skipJob.Spec.Container.Image,
	}
}
