package common

import (
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func DoNotRequeue() (reconcile.Result, error) {
	return reconcile.Result{}, nil
}

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
