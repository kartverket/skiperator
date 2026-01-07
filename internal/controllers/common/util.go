package common

import (
	"fmt"
	"reflect"
	"slices"
	"strings"

	"github.com/chmike/domain"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/api/v1alpha1/podtypes"
	"github.com/kartverket/skiperator/pkg/metrics/usage"
	"github.com/r3labs/diff/v3"
	corev1 "k8s.io/api/core/v1"
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
		usage.ExposeIgnoreResource(obj, 1)
		return false
	}
	return true
}

func IsNamespaceTerminating(namespace *corev1.Namespace) bool {
	return namespace.Status.Phase == corev1.NamespaceTerminating
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

		if err := domain.Check(normalizedHost); err != nil {
			return false
		}
	}

	return true
}

func GetInternalRulesCondition(obj skiperatorv1alpha1.SKIPObject, status metav1.ConditionStatus) metav1.Condition {
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

func GetExternalRulesCondition(obj skiperatorv1alpha1.SKIPObject, status metav1.ConditionStatus) metav1.Condition {
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
