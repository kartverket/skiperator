package common

import (
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// SkiperatorStatus
//
// A status field shown on a Skiperator resource which contains information regarding deployment of the resource.
// +kubebuilder:object:generate=true
type SkiperatorStatus struct {
	Summary            Status             `json:"summary"`
	SubResources       map[string]Status  `json:"subresources"`
	Conditions         []metav1.Condition `json:"conditions"`
	MigrationStartedAt *metav1.Time       `json:"migrationStartedAt,omitempty"`
	// Indicates if access policies are valid
	AccessPolicies StatusNames `json:"accessPolicies"`
}

// Status
//
// +kubebuilder:object:generate=true
type Status struct {
	// +kubebuilder:default=Pending
	Status StatusNames `json:"status"`
	// +kubebuilder:default="Resource accepted by Kubernetes. Waiting for Skiperator to become aware of the resource and start processing."
	Message   string `json:"message"`
	TimeStamp string `json:"timestamp"`
}

type StatusNames string

const (
	SYNCED        StatusNames = "Synced"
	PROGRESSING   StatusNames = "Progressing"
	ERROR         StatusNames = "Error"
	PENDING       StatusNames = "Pending"
	READY         StatusNames = "Ready"
	INVALIDCONFIG StatusNames = "InvalidConfig"

	ReadyConditionType                = "Ready"
	StandardRoutingReadyConditionType = "StandardRoutingReady"
	LegacyRoutingActiveConditionType  = "LegacyRoutingActive"
	SharedRoutingResourcesType        = "SharedRoutingResources"

	// MigrationStalledReason is the condition reason written when a Gateway API
	// migration has kept legacy routing active past the deadline. Shared so the
	// usage metrics package counts exactly the reason gwapi writes.
	MigrationStalledReason = "MigrationStalled"
)

func (s *SkiperatorStatus) SetSummaryPending() {
	s.Summary.Status = PENDING
	s.Summary.Message = "Awaiting first reconcile"
	s.Summary.TimeStamp = metav1.Now().String()
	if s.Conditions == nil {
		s.Conditions = make([]metav1.Condition, 0)
	}
}

func (s *SkiperatorStatus) SetSummarySynced() {
	s.Summary.Status = SYNCED
	s.Summary.Message = "All subresources synced"
	s.Summary.TimeStamp = metav1.Now().String()
	if s.Conditions == nil {
		s.Conditions = make([]metav1.Condition, 0)
	}
}

func (s *SkiperatorStatus) SetSummaryProgressing() {
	s.Summary.Status = PROGRESSING
	s.Summary.Message = "Resource is progressing"
	s.Summary.TimeStamp = metav1.Now().String()
	if s.Conditions == nil {
		s.Conditions = make([]metav1.Condition, 0)
	}
	s.SubResources = make(map[string]Status)
	s.AccessPolicies = PENDING
}

// SetSummaryProgressingMessage marks the summary as progressing with a custom
// message, without resetting subresource state. Use after subresources are
// reconciled but an asynchronous dependency (e.g. standard routing readiness)
// is still pending.
func (s *SkiperatorStatus) SetSummaryProgressingMessage(message string) {
	s.Summary.Status = PROGRESSING
	s.Summary.Message = message
	s.Summary.TimeStamp = metav1.Now().String()
	if s.Conditions == nil {
		s.Conditions = make([]metav1.Condition, 0)
	}
}

func (s *SkiperatorStatus) SetSummaryError(errorMsg string) {
	s.Summary.Status = ERROR
	s.Summary.Message = errorMsg
	s.Summary.TimeStamp = metav1.Now().String()
	if s.Conditions == nil {
		s.Conditions = make([]metav1.Condition, 0)
	}
}

func (s *SkiperatorStatus) SetReadyCondition(status metav1.ConditionStatus, observedGeneration int64, reason string, message string) {
	meta.SetStatusCondition(&s.Conditions, metav1.Condition{
		Type:               ReadyConditionType,
		Status:             status,
		ObservedGeneration: observedGeneration,
		LastTransitionTime: metav1.Now(),
		Reason:             reason,
		Message:            message,
	})
}

func (s *SkiperatorStatus) SetStandardRoutingReadyCondition(status metav1.ConditionStatus, observedGeneration int64, reason string, message string) {
	meta.SetStatusCondition(&s.Conditions, metav1.Condition{
		Type:               StandardRoutingReadyConditionType,
		Status:             status,
		ObservedGeneration: observedGeneration,
		LastTransitionTime: metav1.Now(),
		Reason:             reason,
		Message:            message,
	})
}

func (s *SkiperatorStatus) SetLegacyRoutingActiveCondition(status metav1.ConditionStatus, observedGeneration int64, reason string, message string) {
	meta.SetStatusCondition(&s.Conditions, metav1.Condition{
		Type:               LegacyRoutingActiveConditionType,
		Status:             status,
		ObservedGeneration: observedGeneration,
		LastTransitionTime: metav1.Now(),
		Reason:             reason,
		Message:            message,
	})
}

// SetSharedRoutingResourcesCondition records whether a Routing depends on
// shared Gateway API listener, redirect, and certificate resources.
func (s *SkiperatorStatus) SetSharedRoutingResourcesCondition(status metav1.ConditionStatus, observedGeneration int64, reason string, message string) {
	meta.SetStatusCondition(&s.Conditions, metav1.Condition{
		Type:               SharedRoutingResourcesType,
		Status:             status,
		ObservedGeneration: observedGeneration,
		LastTransitionTime: metav1.Now(),
		Reason:             reason,
		Message:            message,
	})
}

func (s *SkiperatorStatus) AddSubResourceStatus(object client.Object, message string, status StatusNames) {
	if s.SubResources == nil {
		s.SubResources = map[string]Status{}
	}
	kind := object.GetObjectKind().GroupVersionKind().Kind
	key := kind + "[" + object.GetName() + "]"
	s.SubResources[key] = Status{
		Status:    status,
		Message:   kind + " " + message,
		TimeStamp: metav1.Now().String(),
	}

}
