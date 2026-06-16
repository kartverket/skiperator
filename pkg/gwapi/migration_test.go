package gwapi

import (
	"testing"
	"time"

	commontypes "github.com/kartverket/skiperator/api/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestUpdateRoutingStatusTracksLegacyRoutingAndMigration(t *testing.T) {
	status := &commontypes.SkiperatorStatus{}

	events := UpdateRoutingStatus(status, 1, RoutingStateResult{
		GenerateLegacyRouting: true,
		Readiness:             Readiness{Message: "waiting"},
		state:                 routingStateMigratingWithFallback,
	})

	require.Len(t, events, 1)
	assert.Equal(t, "GatewayAPIMigrationStarted", events[0].Reason)
	assert.False(t, events[0].Warning)
	assert.NotNil(t, status.MigrationStartedAt)
	standard := meta.FindStatusCondition(status.Conditions, commontypes.StandardRoutingReadyConditionType)
	require.NotNil(t, standard)
	assert.Equal(t, metav1.ConditionFalse, standard.Status)
	assert.Equal(t, "StandardRoutingNotReady", standard.Reason)
	legacy := meta.FindStatusCondition(status.Conditions, commontypes.LegacyRoutingActiveConditionType)
	require.NotNil(t, legacy)
	assert.Equal(t, metav1.ConditionTrue, legacy.Status)
}

func TestUpdateRoutingStatusMarksStalledMigration(t *testing.T) {
	started := metav1.NewTime(time.Now().Add(-11 * time.Minute))
	status := &commontypes.SkiperatorStatus{MigrationStartedAt: &started}

	events := UpdateRoutingStatus(status, 1, RoutingStateResult{
		GenerateLegacyRouting: true,
		Readiness:             Readiness{Message: "waiting"},
		state:                 routingStateMigrationStalled,
	})

	require.Len(t, events, 1)
	assert.True(t, events[0].Warning)
	assert.Equal(t, "GatewayAPIMigrationStalled", events[0].Reason)
	standard := meta.FindStatusCondition(status.Conditions, commontypes.StandardRoutingReadyConditionType)
	require.NotNil(t, standard)
	assert.Equal(t, "MigrationStalled", standard.Reason)
}

func TestUpdateRoutingStatusClearsMigrationWhenReady(t *testing.T) {
	started := metav1.NewTime(time.Now().Add(-11 * time.Minute))
	status := &commontypes.SkiperatorStatus{MigrationStartedAt: &started}

	events := UpdateRoutingStatus(status, 1, RoutingStateResult{
		GenerateLegacyRouting: false,
		Readiness:             Readiness{Ready: true, Message: "ready"},
		state:                 routingStateStandardOnly,
	})

	require.Len(t, events, 1)
	assert.Equal(t, "GatewayAPIMigrationFinished", events[0].Reason)
	assert.False(t, events[0].Warning)
	assert.Nil(t, status.MigrationStartedAt)
	legacy := meta.FindStatusCondition(status.Conditions, commontypes.LegacyRoutingActiveConditionType)
	require.NotNil(t, legacy)
	assert.Equal(t, metav1.ConditionFalse, legacy.Status)
}

func TestUpdateRoutingStatusSkipsMigrationEventsForGreenfield(t *testing.T) {
	status := &commontypes.SkiperatorStatus{}

	events := UpdateRoutingStatus(status, 1, RoutingStateResult{
		GenerateLegacyRouting: false,
		Readiness:             Readiness{Message: "waiting"},
		state:                 routingStateGreenfieldPending,
	})

	assert.Empty(t, events)
	assert.Nil(t, status.MigrationStartedAt)
}

func TestUpdateRoutingStatusDoesNotStartMigrationForLegacyOnly(t *testing.T) {
	status := &commontypes.SkiperatorStatus{}

	events := UpdateRoutingStatus(status, 1, RoutingStateResult{
		GenerateLegacyRouting: true,
		Readiness:             Readiness{Message: "using only official Istio APIs"},
		state:                 routingStateLegacyOnly,
	})

	assert.Empty(t, events)
	assert.Nil(t, status.MigrationStartedAt)
}

func TestUpdateRoutingStatusEmitsStalledEventOnce(t *testing.T) {
	started := metav1.NewTime(time.Now().Add(-11 * time.Minute))
	status := &commontypes.SkiperatorStatus{
		MigrationStartedAt: &started,
		Conditions: []metav1.Condition{
			{
				Type:   commontypes.StandardRoutingReadyConditionType,
				Status: metav1.ConditionFalse,
				Reason: "MigrationStalled",
			},
		},
	}

	events := UpdateRoutingStatus(status, 1, RoutingStateResult{
		GenerateLegacyRouting: true,
		Readiness:             Readiness{Message: "waiting"},
		state:                 routingStateMigrationStalled,
	})

	assert.Empty(t, events)
}
