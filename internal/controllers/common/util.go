package common

import (
	"fmt"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/api/v1alpha1/podtypes"
	"github.com/r3labs/diff/v3"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
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
	return labels["skiperator.kartverket.no/ignore"] != "true"
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

func GetObjectDiff[T any](a T, b T) (diff.Changelog, error) {
	aKind := reflect.ValueOf(a).Kind()
	bKind := reflect.ValueOf(b).Kind()
	if aKind != bKind {
		return nil, fmt.Errorf("The objects to compare are not the same, found obj1: %v, obj2: %v\n", aKind, bKind)
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
