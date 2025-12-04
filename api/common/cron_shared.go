// +k8s:deepcopy-gen=package
package common

import (
	batchv1 "k8s.io/api/batch/v1"
)

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
