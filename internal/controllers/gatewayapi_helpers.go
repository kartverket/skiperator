package controllers

import (
	"context"

	commontypes "github.com/kartverket/skiperator/api/common"
	controllercommon "github.com/kartverket/skiperator/internal/controllers/common"
	"github.com/kartverket/skiperator/pkg/gwapi"
	"github.com/kartverket/skiperator/pkg/log"
	"k8s.io/apimachinery/pkg/runtime"
)

// This file keeps Gateway API controller glue local to the controllers package.
// Application and Routing share this path; SKIPJob does not satisfy
// gatewayAPIRoutable and cannot call these helpers.

type gatewayAPIRoutable interface {
	commontypes.SKIPObject
	gwapi.Routable
}

func checkGatewayAPIPrerequisites(ctx context.Context, r *controllercommon.ReconcilerBase, obj gatewayAPIRoutable, istioEnabled bool, logger log.Logger) bool {
	if err := r.ValidateIstioEnabledForGatewayAPI(obj.UsesStandardRouting(), istioEnabled, obj.GetNamespace()); err != nil {
		logger.Error(err, "gateway api requires istio revision label")
		r.SetErrorState(ctx, obj, err, "gateway api requires istio revision label", "NamespaceMissingIstioInjection")
		return true
	}

	if err := gwapi.ValidateConflicts(ctx, r.GetClient(), obj); err != nil {
		logger.Error(err, "gateway api conflict")
		r.SetErrorState(ctx, obj, err, "gateway api conflict", "GatewayAPIConflict")
		return true
	}

	return false
}

// finalizeRoutingStatus writes the routing-derived Ready/StandardRoutingReady/
// LegacyRoutingActive conditions and the summary, shared by Application and
// Routing so the two controllers cannot drift. Standard routing's Ready comes
// from Gateway API readiness; legacy routing reports reconciled and clears any
// stale Gateway API conditions (and migration clock) from a prior attempt.
func finalizeRoutingStatus(r *controllercommon.ReconcilerBase, obj gatewayAPIRoutable, routingState gwapi.RoutingStateResult, message string) {
	if obj.UsesStandardRouting() {
		emitMigrationEvents(r, obj, gwapi.UpdateRoutingStatus(obj.GetStatus(), obj.GetGeneration(), routingState))
		if routingState.Readiness.Ready {
			r.EmitNormalEvent(obj, "ReconcileEndSuccess", message)
			obj.GetStatus().SetSummarySynced()
		} else {
			// Subresources reconciled, but standard routing is not ready yet:
			// do not report Synced while the object is not routable.
			obj.GetStatus().SetSummaryProgressingMessage("Subresources reconciled, standard routing migration in progress")
		}
		return
	}

	controllercommon.ClearGatewayAPIConditions(obj)
	controllercommon.SetReadyReconciled(obj, message)
	r.EmitNormalEvent(obj, "ReconcileEndSuccess", message)
	obj.GetStatus().SetSummarySynced()
}

func emitMigrationEvents(r *controllercommon.ReconcilerBase, obj runtime.Object, events []gwapi.MigrationEvent) {
	for _, event := range events {
		if event.Warning {
			r.EmitWarningEvent(obj, event.Reason, event.Message)
			continue
		}
		r.EmitNormalEvent(obj, event.Reason, event.Message)
	}
}
