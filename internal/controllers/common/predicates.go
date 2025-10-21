package common

import (
	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/util"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

var DefaultPredicate = predicate.Funcs{
	CreateFunc: func(e event.CreateEvent) bool {
		switch e.Object.(type) {
		case *skiperatorv1alpha1.Application,
			*corev1.Secret,
			*certmanagerv1.Certificate:
			return true
		default:
			return false
		}
	},
}

var DeploymentPredicate = predicate.Funcs{
	UpdateFunc: func(e event.UpdateEvent) bool {
		if e.ObjectOld == nil || e.ObjectNew == nil {
			return true
		}

		oldCopy := e.ObjectOld.DeepCopyObject()
		newCopy := e.ObjectNew.DeepCopyObject()

		oldDep := oldCopy.(*appsv1.Deployment)
		newDep := newCopy.(*appsv1.Deployment)

		// HPA Should not trigger reconciles
		// Manually adjusting replicas will no longer trigger reconciles, but this saves us 1 full reconcile
		newDep.Spec.Replicas = oldDep.Spec.Replicas
		oldHash := util.GetHashForStructs([]interface{}{&oldDep.Spec, &oldDep.Labels})
		newHash := util.GetHashForStructs([]interface{}{&newDep.Spec, &newDep.Labels})

		if oldHash != newHash {
			return true
		}
		return false
	},
}
