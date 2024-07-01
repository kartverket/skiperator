package routingcontroller

import (
	"context"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	ConditionStatusTrue    = "True"
	ConditionStatusFalse   = "False"
	ConditionStatusUnknown = "Unknown"

	ConditionTypeCertificateSynced     = "CertificateSynced"
	ConditionReasonCertificateSynced   = "CertificateSynced"
	ConditionMessageCertificateSynced  = "Certificate has been synced"
	ConditionMessageCertificateSkipped = "Certificate has been skipped (custom certificate secret in use)"

	ConditionTypeGatewaySynced                     = "GatewaySynced"
	ConditionReasonGatewaySynced                   = "GatewaySynced"
	ConditionMessageGatewaySynced                  = "Gateway has been synced"
	ConditionMessageGatewaySyncedCustomCertificate = "Gateway has been synced (using a custom certificate)"

	ConditionTypeVirtualServiceSynced    = "VirtualServiceSynced"
	CoditionReasonVirtualServiceSynced   = "VirtualServiceSynced"
	ConditionMessageVirtualServiceSynced = "VirtualService has been synced"

	ConditionTypeNetworkPolicySynced    = "NetworkPolicySynced"
	ConditionReasonNetworkPolicySynced  = "NetworkPolicySynced"
	ConditionMessageNetworkPolicySynced = "NetworkPolicy has been synced"
)

func (r *RoutingReconciler) setConditionCertificateSynced(ctx context.Context, routing *skiperatorv1alpha1.Routing, status metav1.ConditionStatus, message string) error {
	if !r.containsCondition(ctx, routing, ConditionReasonCertificateSynced) {
		return util.AppendCondition(ctx, r.GetClient(), routing, ConditionTypeCertificateSynced, status,
			ConditionReasonCertificateSynced, message)
	} else {
		currentStatus := r.getConditionStatus(ctx, routing, ConditionTypeCertificateSynced)
		if currentStatus != status {
			r.deleteCondition(ctx, routing, ConditionTypeCertificateSynced, ConditionReasonCertificateSynced)
			return util.AppendCondition(ctx, r.GetClient(), routing, ConditionTypeCertificateSynced, status,
				ConditionReasonCertificateSynced, message)
		}
	}
	return nil
}

func (r *RoutingReconciler) setConditionGatewaySynced(ctx context.Context, routing *skiperatorv1alpha1.Routing, status metav1.ConditionStatus, message string) error {
	if !r.containsCondition(ctx, routing, ConditionReasonGatewaySynced) {
		return util.AppendCondition(ctx, r.GetClient(), routing, ConditionTypeGatewaySynced, status,
			ConditionReasonGatewaySynced, message)
	} else {
		currentStatus := r.getConditionStatus(ctx, routing, ConditionTypeGatewaySynced)
		if currentStatus != status {
			r.deleteCondition(ctx, routing, ConditionTypeGatewaySynced, ConditionReasonGatewaySynced)
			return util.AppendCondition(ctx, r.GetClient(), routing, ConditionTypeGatewaySynced, status,
				ConditionReasonGatewaySynced, message)
		}
	}
	return nil
}

func (r *RoutingReconciler) setConditionVirtualServiceSynced(ctx context.Context, routing *skiperatorv1alpha1.Routing, status metav1.ConditionStatus, message string) error {
	if !r.containsCondition(ctx, routing, CoditionReasonVirtualServiceSynced) {
		return util.AppendCondition(ctx, r.GetClient(), routing, ConditionTypeVirtualServiceSynced, ConditionStatusTrue,
			CoditionReasonVirtualServiceSynced, ConditionMessageVirtualServiceSynced)
	} else {
		currentStatus := r.getConditionStatus(ctx, routing, ConditionTypeVirtualServiceSynced)
		if currentStatus != status {
			r.deleteCondition(ctx, routing, ConditionTypeVirtualServiceSynced, CoditionReasonVirtualServiceSynced)
			return util.AppendCondition(ctx, r.GetClient(), routing, ConditionTypeVirtualServiceSynced, status,
				CoditionReasonVirtualServiceSynced, message)
		}
	}
	return nil
}

func (r *RoutingReconciler) setConditionNetworkPolicySynced(ctx context.Context, routing *skiperatorv1alpha1.Routing, status metav1.ConditionStatus, message string) error {
	if !r.containsCondition(ctx, routing, ConditionReasonNetworkPolicySynced) {
		return util.AppendCondition(ctx, r.GetClient(), routing, ConditionTypeNetworkPolicySynced, ConditionStatusTrue,
			ConditionReasonNetworkPolicySynced, ConditionMessageNetworkPolicySynced)
	} else {
		currentStatus := r.getConditionStatus(ctx, routing, ConditionTypeNetworkPolicySynced)
		if currentStatus != status {
			r.deleteCondition(ctx, routing, ConditionTypeNetworkPolicySynced, ConditionReasonNetworkPolicySynced)
			return util.AppendCondition(ctx, r.GetClient(), routing, ConditionTypeNetworkPolicySynced, status,
				ConditionReasonNetworkPolicySynced, message)
		}
	}
	return nil
}

func (r *RoutingReconciler) getConditionStatus(ctx context.Context, routing *skiperatorv1alpha1.Routing, typeName string) metav1.ConditionStatus {

	var output metav1.ConditionStatus = ConditionStatusUnknown
	for _, condition := range routing.Status.Conditions {
		if condition.Type == typeName {
			return condition.Status
		}
	}
	return output
}

func (r *RoutingReconciler) deleteCondition(ctx context.Context, routing *skiperatorv1alpha1.Routing, typeName string, reason string) error {
	logger := log.FromContext(ctx)
	var newConditions = make([]metav1.Condition, 0)
	for _, condition := range routing.Status.Conditions {
		if condition.Type != typeName && condition.Reason != reason {
			newConditions = append(newConditions, condition)
		}
	}
	routing.Status.Conditions = newConditions

	err := r.GetClient().Status().Update(ctx, routing)
	if err != nil {
		logger.Info("Routing resource status update failed")
	}

	return nil
}

func (r *RoutingReconciler) containsCondition(ctx context.Context, routing *skiperatorv1alpha1.Routing, reason string) bool {
	output := false
	for _, condition := range routing.Status.Conditions {
		if condition.Reason == reason {
			output = true
		}
	}
	return output
}
