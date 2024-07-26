package v1alpha1

import corev1 "k8s.io/api/core/v1"

/*
 *  SKIPNamespace is a wrapper for the kubernetes namespace resource, so we can utilize the SKIPObject interface
 */

type SKIPNamespace struct {
	*corev1.Namespace
}

func (n SKIPNamespace) GetStatus() *SkiperatorStatus {
	return &SkiperatorStatus{}
}
