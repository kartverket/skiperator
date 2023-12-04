package v1alpha1

import (
	"github.com/kartverket/skiperator/api/v1alpha1/podtypes"
	"github.com/kartverket/skiperator/pkg/util"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
// SKIPJob is the Schema for the skipjobs API
type SKIPJob struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	//+kubebuilder:validation:Required
	Spec SKIPJobSpec `json:"spec"`

	//+kubebuilder:validation:Optional
	Status SKIPJobStatus `json:"status"`
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
// A SKIPJob is either defined as a one-off or a scheduled job. If the Cron field is set for SKIPJob, it may not be removed. If the Cron field is unset, it may not be added.
// The Container field of a SKIPJob is only mutable if the Cron field is set. If unset, you must delete your SKIPJob to change container settings.
//
// +kubebuilder:validation:XValidation:rule="(has(oldSelf.cron) && has(self.cron)) || (!has(oldSelf.cron) && !has(self.cron))", message="After creation of a SKIPJob you may not remove the Cron field if it was previously present, or add it if it was previously omitted. Please delete the SKIPJob to change its nature from a one-off/scheduled job."
// +kubebuilder:validation:XValidation:rule="((!has(self.cron) && (oldSelf.container == self.container)) || has(self.cron))", message="The field Container is immutable for one-off jobs. Please delete your SKIPJob to change the containers settings."
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
	// Once set, you may not change Container without deleting your current SKIPJob
	//
	// +kubebuilder:validation:Required
	Container ContainerSettings `json:"container"`

	// Prometheus settings for pod running in job. Fields are identical to Application and if set,
	// a monitorngs object is created.
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
	return util.ResourceNameWithKindPostfix(skipJob.Name, skipJob.Kind)
}
