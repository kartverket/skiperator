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

// DefaultPredicate limits Create events to the types that skiperator reacts to.
// The subresources that skiperator creates itself do not start a second
// reconcile. corev1.Service is in the list because the ports of an outbound
// access policy come from the Service of the target application. A new Service
// is therefore the signal that unblocks every object that references it (see
// OutboundTargets).
//
// The controller installs this predicate through WithEventFilter, so it applies
// to every watch. A type must appear in the list below, or its Create events
// never reach the reconciler. A predicate per watch through
// builder.WithPredicates is more precise, but then every Owns() call needs a
// predicate of its own.
var DefaultPredicate = predicate.Funcs{
	CreateFunc: func(e event.CreateEvent) bool {
		switch e.Object.(type) {
		case *skiperatorv1alpha1.Application,
			*corev1.Secret,
			*corev1.Service,
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
		oldHash := util.GetHashForStructs([]any{&oldDep.Spec, &oldDep.Labels})
		newHash := util.GetHashForStructs([]any{&newDep.Spec, &newDep.Labels})

		return oldHash != newHash
	},
}

var StatefulSetPredicate = predicate.Funcs{
	UpdateFunc: func(e event.UpdateEvent) bool {
		if e.ObjectOld == nil || e.ObjectNew == nil {
			return true
		}

		oldCopy := e.ObjectOld.DeepCopyObject()
		newCopy := e.ObjectNew.DeepCopyObject()

		oldSts := oldCopy.(*appsv1.StatefulSet)
		newSts := newCopy.(*appsv1.StatefulSet)

		newSts.Spec.Replicas = oldSts.Spec.Replicas
		oldHash := util.GetHashForStructs([]any{&oldSts.Spec, &oldSts.Labels})
		newHash := util.GetHashForStructs([]any{&newSts.Spec, &newSts.Labels})

		return oldHash != newHash
	},
}
