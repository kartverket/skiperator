package common

import (
	"fmt"
	"reflect"
	"slices"
	"strings"
	"time"

	"github.com/chmike/domain"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/kartverket/skiperator/api/common"
	"github.com/kartverket/skiperator/api/common/podtypes"
	"github.com/kartverket/skiperator/pkg/mesh"
	"github.com/kartverket/skiperator/pkg/metrics/usage"
	"github.com/r3labs/diff/v3"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func DoNotRequeue() (reconcile.Result, error) {
	return reconcile.Result{}, nil
}

// TODO: exponential backoff
func RequeueWithError(err error) (reconcile.Result, error) {
	return reconcile.Result{}, err
}

func ShouldReconcile(obj client.Object) bool {
	labels := obj.GetLabels()
	if labels["skiperator.kartverket.no/ignore"] == "true" {
		// Expose metrics for ignored resource
		usage.ExposeIgnoredResource(obj)
		return false
	}

	usage.RemoveIgnoredResource(obj)
	return true
}

func IsNamespaceTerminating(namespace *corev1.Namespace) bool {
	return namespace.Status.Phase == corev1.NamespaceTerminating
}

// OutboundTargetIndex is the field index that Application and SKIPJob
// register on the applications their outbound rules target. The index maps a
// changed Service back to its dependents. Without it, each Service event must
// list every object.
//
// Inbound rules need no index. They compile to label selectors that Kubernetes
// evaluates, so the operator has nothing to recompute when a peer appears.
const OutboundTargetIndex = "accessPolicy.outbound.rules.application"

// AccessPolicyRequeueDelay is how often an object with unresolved internal
// rules reconciles again. The Service watch on each controller covers a target
// Service that appears or is deleted. This delay covers the two
// cases that no watch reports. A target Service can gain ports, and a namespace
// can gain a label that a rule selects on.
const AccessPolicyRequeueDelay = 5 * time.Minute

// OutboundTargets lists the applications that the outbound rules target,
// for use as index values. The namespace is not part of the key. Rules can
// select namespaces by label, and that set of namespaces needs a lookup. A key
// of the name alone selects too many objects, but it never misses a dependent.
// A reconcile is idempotent, so a redundant one costs less than the lookup.
//
// This function takes the access policy and not a client.Object, so that it
// stays pure and has no panic path of its own. An index function runs inside the informer,
// where a panic stops the process. The spec accessors on Application and
// SKIPJob read fields that the API server defaults. An object built in memory
// does not have those defaults.
func OutboundTargets(accessPolicy *podtypes.AccessPolicy) []string {
	if accessPolicy == nil || accessPolicy.Outbound == nil {
		return nil
	}

	targets := make([]string, len(accessPolicy.Outbound.Rules))
	for i, rule := range accessPolicy.Outbound.Rules {
		targets[i] = rule.Application
	}
	return targets
}

func IsInternalRulesValid(accessPolicy *podtypes.AccessPolicy) bool {
	if accessPolicy == nil || accessPolicy.Outbound == nil {
		return true
	}

	for _, rule := range accessPolicy.Outbound.Rules {
		if len(rule.Ports) == 0 {
			return false
		}
	}

	return true
}

func IsExternalRulesValid(accessPolicy *podtypes.AccessPolicy) bool {
	if accessPolicy == nil || accessPolicy.Outbound == nil {
		return true
	}

	seenHosts := []string{}
	for _, rule := range accessPolicy.Outbound.External {
		if len(rule.Host) == 0 {
			return false
		}

		normalizedHost := strings.ToLower(rule.Host)
		if slices.Contains(seenHosts, normalizedHost) {
			return false
		}
		seenHosts = append(seenHosts, normalizedHost)

		if normalizedHost == rule.Ip {
			return true
		}

		if err := domain.Check(normalizeWildcardHost(normalizedHost)); err != nil {
			return false
		}
	}

	return true
}

func normalizeWildcardHost(host string) string {
	return strings.TrimPrefix(host, "*.")
}

func GetInternalRulesCondition(obj common.SKIPObject, status metav1.ConditionStatus) metav1.Condition {
	message := "Internal rules are valid"
	if status == metav1.ConditionFalse {
		message = "Internal rules are invalid, applications or namespaces defined might not exist or have invalid ports"
	}
	return metav1.Condition{
		Type:               "InternalRulesValid",
		Status:             status,
		ObservedGeneration: obj.GetGeneration(),
		LastTransitionTime: metav1.Now(),
		Reason:             "ApplicationReconciled",
		Message:            message,
	}
}

func GetExternalRulesCondition(obj common.SKIPObject, status metav1.ConditionStatus) metav1.Condition {
	message := "External rules are valid"
	if status == metav1.ConditionFalse {
		message = "External rules are invalid – hostname may be empty or duplicate, or the hostname may not be a valid DNS name"
	}
	return metav1.Condition{
		Type:               "ExternalRulesValid",
		Status:             status,
		ObservedGeneration: obj.GetGeneration(),
		LastTransitionTime: metav1.Now(),
		Reason:             "ApplicationReconciled",
		Message:            message,
	}
}

// SetInternalRulesCondition and SetExternalRulesCondition merge the rules
// conditions in place via meta.SetStatusCondition, which preserves
// LastTransitionTime when the status does not change (unlike rebuilding the
// condition slice from scratch every reconcile).
func SetInternalRulesCondition(obj common.SKIPObject, status metav1.ConditionStatus) {
	meta.SetStatusCondition(&obj.GetStatus().Conditions, GetInternalRulesCondition(obj, status))
}

func SetExternalRulesCondition(obj common.SKIPObject, status metav1.ConditionStatus) {
	meta.SetStatusCondition(&obj.GetStatus().Conditions, GetExternalRulesCondition(obj, status))
}

func SetReadyInvalidConfig(obj common.SKIPObject, message string) {
	obj.GetStatus().SetReadyCondition(metav1.ConditionFalse, obj.GetGeneration(), "InvalidConfig", message)
}

func SetReadyReconciled(obj common.SKIPObject, message string) {
	obj.GetStatus().SetReadyCondition(metav1.ConditionTrue, obj.GetGeneration(), "Reconciled", message)
}

// SetSharedRoutingResourcesActive marks that the object contributes to shared
// Gateway API resources, which live outside its own namespace.
func SetSharedRoutingResourcesActive(obj common.SKIPObject) {
	obj.GetStatus().SetSharedRoutingResourcesCondition(
		metav1.ConditionTrue,
		obj.GetGeneration(),
		"SharedRoutingResourcesActive",
		fmt.Sprintf("Routing uses shared Gateway API resources in %s", mesh.GatewayNamespace),
	)
}

// SetRoutePathConflict marks that this object's paths overlap another accepted
// route on the same hostname. Gateway API has already decided which route
// answers a request, so the condition reports the overlap instead of blocking
// the reconcile.
func SetRoutePathConflict(obj common.SKIPObject, message string) {
	obj.GetStatus().SetRoutePathConflictCondition(
		metav1.ConditionTrue,
		obj.GetGeneration(),
		"OverlappingPathPrefix",
		message,
	)
}

// ClearRoutePathConflictCondition drops the condition when no overlap remains.
func ClearRoutePathConflictCondition(obj common.SKIPObject) {
	meta.RemoveStatusCondition(&obj.GetStatus().Conditions, common.RoutePathConflictType)
}

// ClearSharedRoutingResourcesCondition drops the condition, so a standalone
// object does not carry one.
func ClearSharedRoutingResourcesCondition(obj common.SKIPObject) {
	meta.RemoveStatusCondition(&obj.GetStatus().Conditions, common.SharedRoutingResourcesType)
}

func ClearGatewayAPIConditions(obj common.SKIPObject) {
	meta.RemoveStatusCondition(&obj.GetStatus().Conditions, common.StandardRoutingReadyConditionType)
	meta.RemoveStatusCondition(&obj.GetStatus().Conditions, common.LegacyRoutingActiveConditionType)
	meta.RemoveStatusCondition(&obj.GetStatus().Conditions, common.SharedRoutingResourcesType)
	// Drop the migration clock too, so switching back to legacy does not leave
	// a stale start time that mis-seeds a future migration as already stalled.
	obj.GetStatus().MigrationStartedAt = nil
}

func GetObjectDiff[T any](a T, b T) (diff.Changelog, error) {
	aKind := reflect.ValueOf(a).Kind()
	bKind := reflect.ValueOf(b).Kind()
	if aKind != bKind {
		return nil, fmt.Errorf("the objects to compare are not the same, found obj1: %v, obj2: %v", aKind, bKind)
	}
	changelog, err := diff.Diff(a, b)

	changelog = filterOutStatusTimestamps(changelog)

	if len(changelog) == 0 {
		return nil, err
	}

	return changelog, nil
}

func filterOutStatusTimestamps(changelog diff.Changelog) diff.Changelog {
	changelog = changelog.FilterOut([]string{"Summary", "TimeStamp"})
	changelog = changelog.FilterOut([]string{"Conditions", ".*", "LastTransitionTime"})
	changelog = changelog.FilterOut([]string{"SubResources", ".*", "TimeStamp"})
	return changelog
}

func ValidateContainerImageString(obj common.SKIPObject) error {
	return ValidateImageString(obj.GetCommonSpec().Image)
}

func ValidateImageString(image string) error {
	_, err := name.ParseReference(image)
	if err != nil {
		return err
	}
	return nil
}
