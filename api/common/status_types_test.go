package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestSetReadyCondition(t *testing.T) {
	status := &SkiperatorStatus{}

	status.SetReadyCondition(metav1.ConditionFalse, 3, "ValidationFailed", "not ready")

	requirement := meta.FindStatusCondition(status.Conditions, ReadyConditionType)
	if assert.NotNil(t, requirement) {
		assert.Equal(t, metav1.ConditionFalse, requirement.Status)
		assert.Equal(t, int64(3), requirement.ObservedGeneration)
		assert.Equal(t, "ValidationFailed", requirement.Reason)
		assert.Equal(t, "not ready", requirement.Message)
	}
}

func TestSetStandardRoutingReadyCondition(t *testing.T) {
	status := &SkiperatorStatus{}

	status.SetStandardRoutingReadyCondition(metav1.ConditionTrue, 4, "StandardRoutingReady", "ready")

	requirement := meta.FindStatusCondition(status.Conditions, StandardRoutingReadyConditionType)
	if assert.NotNil(t, requirement) {
		assert.Equal(t, metav1.ConditionTrue, requirement.Status)
		assert.Equal(t, int64(4), requirement.ObservedGeneration)
		assert.Equal(t, "StandardRoutingReady", requirement.Reason)
		assert.Equal(t, "ready", requirement.Message)
	}
}

func TestSetLegacyRoutingActiveCondition(t *testing.T) {
	status := &SkiperatorStatus{}

	status.SetLegacyRoutingActiveCondition(metav1.ConditionTrue, 5, "LegacyRoutingActive", "legacy routing is active")

	requirement := meta.FindStatusCondition(status.Conditions, LegacyRoutingActiveConditionType)
	if assert.NotNil(t, requirement) {
		assert.Equal(t, metav1.ConditionTrue, requirement.Status)
		assert.Equal(t, int64(5), requirement.ObservedGeneration)
		assert.Equal(t, "LegacyRoutingActive", requirement.Reason)
		assert.Equal(t, "legacy routing is active", requirement.Message)
	}
}

func TestSetSharedRoutingResourcesCondition(t *testing.T) {
	status := &SkiperatorStatus{}

	status.SetSharedRoutingResourcesCondition(metav1.ConditionTrue, 6, "SharedRoutingResourcesActive", "shared routing resources are active")

	requirement := meta.FindStatusCondition(status.Conditions, SharedRoutingResourcesType)
	if assert.NotNil(t, requirement) {
		assert.Equal(t, metav1.ConditionTrue, requirement.Status)
		assert.Equal(t, int64(6), requirement.ObservedGeneration)
		assert.Equal(t, "SharedRoutingResourcesActive", requirement.Reason)
		assert.Equal(t, "shared routing resources are active", requirement.Message)
	}
}
