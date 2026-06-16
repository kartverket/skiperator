package gwapi

import (
	"fmt"
	"time"

	commontypes "github.com/kartverket/skiperator/api/common"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	standardRoutingReadyReason    = "StandardRoutingReady"
	standardRoutingNotReadyReason = "StandardRoutingNotReady"
	migrationStalledReason        = "MigrationStalled"
	legacyRoutingActiveReason     = "LegacyRoutingActive"
	legacyRoutingInactiveReason   = "LegacyRoutingInactive"

	migrationStartedEventReason  = "GatewayAPIMigrationStarted"
	migrationFinishedEventReason = "GatewayAPIMigrationFinished"
	migrationStalledEventReason  = "GatewayAPIMigrationStalled"

	migrationStalledAfter = 10 * time.Minute

	routingStateLegacyOnly              routingState = "LegacyOnly"
	routingStateGreenfieldPending       routingState = "GreenfieldPending"
	routingStateMigratingWithFallback   routingState = "MigratingWithFallback"
	routingStateMigrationStalled        routingState = "MigrationStalled"
	routingStateCutoverReadyPruneLegacy routingState = "CutoverReadyPruneLegacy"
	routingStateStandardOnly            routingState = "StandardOnly"
	routingStateInvalid                 routingState = "Invalid"
)

type routingState string

// MigrationEvent describes a Kubernetes Event the reconciler should emit when
// a standard-routing migration starts, finishes, or stalls.
type MigrationEvent struct {
	Warning bool
	Reason  string
	Message string
}

// RoutingStateResult is the state machine output consumed by reconcilers.
//
// State itself is intentionally package-private. Callers should make decisions
// from GenerateLegacyRouting and Readiness, while this package owns how states
// map to Skiperator status conditions and events.
type RoutingStateResult struct {
	GenerateLegacyRouting bool
	Readiness             Readiness
	state                 routingState
}

// determineRoutingState turns provider choice, standard-routing readiness, and
// existing legacy resources into one migration state.
//
// The order matters. A ready standard route with legacy resources means cutover
// is safe and legacy can be pruned. A not-ready standard route with legacy
// resources means fallback must stay. A not-ready standard route with no legacy
// resources is a greenfield rollout, so generating legacy resources would create
// an unintended second routing provider.
func determineRoutingState(usesStandardRouting bool, ready Readiness, legacyRoutingExists bool, migrationStartedAt *metav1.Time) RoutingStateResult {
	if !usesStandardRouting {
		return RoutingStateResult{
			GenerateLegacyRouting: true,
			Readiness:             Readiness{Ready: true, Message: "using only official Istio APIs"},
			state:                 routingStateLegacyOnly,
		}
	}
	if ready.Ready && legacyRoutingExists {
		return RoutingStateResult{
			GenerateLegacyRouting: false,
			Readiness:             ready,
			state:                 routingStateCutoverReadyPruneLegacy,
		}
	}
	if ready.Ready {
		return RoutingStateResult{
			GenerateLegacyRouting: false,
			Readiness:             ready,
			state:                 routingStateStandardOnly,
		}
	}
	if !legacyRoutingExists {
		return RoutingStateResult{
			GenerateLegacyRouting: false,
			Readiness:             ready,
			state:                 routingStateGreenfieldPending,
		}
	}
	if migrationStartedAt != nil && time.Since(migrationStartedAt.Time) > migrationStalledAfter {
		return RoutingStateResult{
			GenerateLegacyRouting: true,
			Readiness:             ready,
			state:                 routingStateMigrationStalled,
		}
	}
	return RoutingStateResult{
		GenerateLegacyRouting: true,
		Readiness:             ready,
		state:                 routingStateMigratingWithFallback,
	}
}

// UpdateRoutingStatus writes Ready, StandardRoutingReady, and
// LegacyRoutingActive conditions from a RoutingStateResult.
//
// It also maintains MigrationStartedAt so stalled migrations are reported once
// they have kept legacy fallback active for too long.
func UpdateRoutingStatus(status *commontypes.SkiperatorStatus, generation int64, state RoutingStateResult) []MigrationEvent {
	previous := status.DeepCopy()

	if state.GenerateLegacyRouting {
		status.SetLegacyRoutingActiveCondition(metav1.ConditionTrue, generation, legacyRoutingActiveReason, "Legacy routing remains active while standard routing is not ready")
	} else {
		status.SetLegacyRoutingActiveCondition(metav1.ConditionFalse, generation, legacyRoutingInactiveReason, "Legacy routing is not active")
	}

	if state.Readiness.Ready {
		status.MigrationStartedAt = nil
		status.SetReadyCondition(metav1.ConditionTrue, generation, "Reconciled", "Standard routing is ready")
		status.SetStandardRoutingReadyCondition(metav1.ConditionTrue, generation, standardRoutingReadyReason, state.Readiness.Message)
		return migrationEvents(previous, status, state.GenerateLegacyRouting)
	}

	if state.state == routingStateMigratingWithFallback && status.MigrationStartedAt == nil {
		status.MigrationStartedAt = new(metav1.Now())
	}
	reason := standardRoutingNotReadyReason
	if state.state == routingStateMigrationStalled {
		reason = migrationStalledReason
	}
	status.SetReadyCondition(metav1.ConditionFalse, generation, reason, state.Readiness.Message)
	status.SetStandardRoutingReadyCondition(metav1.ConditionFalse, generation, reason, state.Readiness.Message)
	return migrationEvents(previous, status, state.GenerateLegacyRouting)
}

func migrationEvents(previous *commontypes.SkiperatorStatus, current *commontypes.SkiperatorStatus, generateLegacyRouting bool) []MigrationEvent {
	events := []MigrationEvent{}
	if generateLegacyRouting && previous.MigrationStartedAt == nil && current.MigrationStartedAt != nil {
		events = append(events, MigrationEvent{
			Reason:  migrationStartedEventReason,
			Message: "Started migrating to standard routing, legacy routing is active",
		})
	}
	if previous.MigrationStartedAt != nil && current.MigrationStartedAt == nil {
		events = append(events, MigrationEvent{
			Reason:  migrationFinishedEventReason,
			Message: "Migration completed successfully, legacy routing has been cleaned up",
		})
	}

	previousStandard := meta.FindStatusCondition(previous.Conditions, commontypes.StandardRoutingReadyConditionType)
	currentStandard := meta.FindStatusCondition(current.Conditions, commontypes.StandardRoutingReadyConditionType)
	if generateLegacyRouting &&
		currentStandard != nil &&
		currentStandard.Reason == migrationStalledReason &&
		(previousStandard == nil || previousStandard.Reason != migrationStalledReason) {
		events = append(events, MigrationEvent{
			Warning: true,
			Reason:  migrationStalledEventReason,
			Message: fmt.Sprintf("Contact SKIP for support - migration has stalled (taken longer than deadline %s): %s", migrationStalledAfter.Round(time.Minute), currentStandard.Message),
		})
	}

	return events
}
