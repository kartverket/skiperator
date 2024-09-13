package v1alpha1

import (
	"dario.cat/mergo"
	"fmt"
	"github.com/kartverket/skiperator/api/v1alpha1/istiotypes"
	"github.com/kartverket/skiperator/api/v1alpha1/podtypes"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
	"time"
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
	IstioSettings *istiotypes.IstioSettings `json:"istioSettings,omitempty"`

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

	// +kubebuilder:validation:Enum=OnFailure;Never
	// +kubebuilder:default="Never"
	// +kubebuilder:validation:Optional
	RestartPolicy *corev1.RestartPolicy `json:"restartPolicy"`

	//+kubebuilder:validation:Optional
	PodSettings *podtypes.PodSettings `json:"podSettings,omitempty"`
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

	// If set to true, this tells Kubernetes to suspend this Job till the field is set to false. If the Job is active while this field is set to false,
	// all running Pods will be terminated.
	//
	//+kubebuilder:validation:Optional
	Suspend *bool `json:"suspend,omitempty"`

	// The number of seconds to wait before removing the Job after it has finished. If unset, Job will not be cleaned up.
	// It is recommended to set this to avoid clutter in your resource tree.
	//
	//+kubebuilder:validation:Optional
	TTLSecondsAfterFinished *int32 `json:"ttlSecondsAfterFinished,omitempty"`
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

	// The time zone name for the given schedule, see https://en.wikipedia.org/wiki/List_of_tz_database_time_zones. If not specified,
	// this will default to the time zone of the cluster.
	//
	// Example: "Europe/Oslo"
	//
	// +kubebuilder:validation:Optional
	TimeZone *string `json:"timeZone,omitempty"`

	// Denotes the deadline in seconds for starting a job on its schedule, if for some reason the Job's controller was not ready upon the scheduled time.
	// If unset, Jobs missing their deadline will be considered failed jobs and will not start.
	//
	//+kubebuilder:validation:Optional
	StartingDeadlineSeconds *int64 `json:"startingDeadlineSeconds,omitempty"`

	// If set to true, this tells Kubernetes to suspend this Job till the field is set to false. If the Job is active while this field is set to true,
	// all running Pods will be terminated.
	//
	//+kubebuilder:validation:Optional
	Suspend *bool `json:"suspend,omitempty"`
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

func (skipJob *SKIPJob) FillDefaultSpec() error {
	defaults := &SKIPJob{
		Spec: SKIPJobSpec{
			Job: &JobSettings{
				TTLSecondsAfterFinished: &DefaultTTLSecondsAfterFinished,
				BackoffLimit:            &DefaultBackoffLimit,
				Suspend:                 &DefaultSuspend,
			},
		},
	}

	if skipJob.Spec.Cron != nil {
		defaults.Spec.Cron = &CronSettings{}
		suspend := false
		defaults.Spec.Cron.Suspend = &suspend
	}

	return mergo.Merge(skipJob, defaults)
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
	}
}
