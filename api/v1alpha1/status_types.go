package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ApplicationStatus
//
// A status field shown on a Skiperator resource which contains information regarding deployment of the resource.
// +kubebuilder:object:generate=true
type SkiperatorStatus struct {
	Summary      Status             `json:"summary"`
	SubResources map[string]Status  `json:"subresources"`
	Conditions   []metav1.Condition `json:"conditions"`
}

// Status
//
// +kubebuilder:object:generate=true
type Status struct {
	// +kubebuilder:default="Synced"
	Status StatusNames `json:"status"`
	// +kubebuilder:default="hello"
	Message string `json:"message"`
	// +kubebuilder:default="hello"
	TimeStamp string `json:"timestamp"`
}

type StatusNames string

const (
	SYNCED      StatusNames = "Synced"
	PROGRESSING StatusNames = "Progressing"
	ERROR       StatusNames = "Error"
	PENDING     StatusNames = "Pending"
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
}

func (s *SkiperatorStatus) SetSummaryError(errorMsg string) {
	s.Summary.Status = ERROR
	s.Summary.Message = errorMsg
	s.Summary.TimeStamp = metav1.Now().String()
	if s.Conditions == nil {
		s.Conditions = make([]metav1.Condition, 0)
	}
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
