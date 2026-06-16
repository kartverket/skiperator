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

func checkGatewayAPIPrerequisites(ctx context.Context, r *controllercommon.ReconcilerBase, obj gatewayAPIRoutable, logger log.Logger) bool {
	if err := r.ValidateIstioEnabledForGatewayAPI(ctx, obj.UsesStandardRouting(), obj.GetNamespace()); err != nil {
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

func emitMigrationEvents(r *controllercommon.ReconcilerBase, obj runtime.Object, events []gwapi.MigrationEvent) {
	for _, event := range events {
		if event.Warning {
			r.EmitWarningEvent(obj, event.Reason, event.Message)
			continue
		}
		r.EmitNormalEvent(obj, event.Reason, event.Message)
	}
}
