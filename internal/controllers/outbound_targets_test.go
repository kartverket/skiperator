package controllers

import (
	"context"
	"testing"

	"github.com/kartverket/skiperator/api/common/podtypes"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	controllercommon "github.com/kartverket/skiperator/internal/controllers/common"
	"github.com/kartverket/skiperator/pkg/resourceschemas"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func applicationTargeting(namespace, name string, targets ...string) *skiperatorv1alpha1.Application {
	rules := make([]podtypes.InternalRule, len(targets))
	for i, target := range targets {
		rules[i] = podtypes.InternalRule{Application: target}
	}
	return &skiperatorv1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: name},
		Spec: skiperatorv1alpha1.ApplicationSpec{
			AccessPolicy: &podtypes.AccessPolicy{
				Outbound: &podtypes.OutboundPolicy{Rules: rules},
			},
		},
	}
}

// The index key is the application name alone. A Service therefore enqueues its
// dependents in every namespace. Objects that do not reference the Service get
// no request.
func TestOutboundTargetRequestsEnqueuesDependentsAcrossNamespaces(t *testing.T) {
	scheme := runtime.NewScheme()
	resourceschemas.AddSchemas(scheme)

	c := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(
			applicationTargeting("team-a", "first", "second"),
			applicationTargeting("team-b", "other", "second", "third"),
			applicationTargeting("team-c", "unrelated", "third"),
			&skiperatorv1alpha1.Application{ObjectMeta: metav1.ObjectMeta{Namespace: "team-d", Name: "no-policy"}},
		).
		WithIndex(&skiperatorv1alpha1.Application{}, controllercommon.OutboundTargetIndex, func(obj client.Object) []string {
			return controllercommon.OutboundTargets(obj.(*skiperatorv1alpha1.Application).Spec.AccessPolicy)
		}).
		Build()

	reconciler := &ApplicationReconciler{
		ReconcilerBase: controllercommon.NewReconcilerBase(c, nil, scheme, nil, nil),
	}

	requests := reconciler.outboundTargetRequests(
		context.Background(),
		&corev1.Service{ObjectMeta: metav1.ObjectMeta{Namespace: "team-x", Name: "second"}},
	)

	assert.ElementsMatch(t, []reconcile.Request{
		{NamespacedName: types.NamespacedName{Namespace: "team-a", Name: "first"}},
		{NamespacedName: types.NamespacedName{Namespace: "team-b", Name: "other"}},
	}, requests)
}
