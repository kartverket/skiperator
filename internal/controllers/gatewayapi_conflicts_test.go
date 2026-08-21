package controllers

import (
	"context"
	"testing"

	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	controllercommon "github.com/kartverket/skiperator/internal/controllers/common"
	"github.com/kartverket/skiperator/pkg/resourceschemas"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestApplicationStandardRoutingRequiresIstioRevision(t *testing.T) {
	scheme := runtime.NewScheme()
	resourceschemas.AddSchemas(scheme)
	namespace := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "team-a"}}
	client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(namespace).Build()
	reconciler := &ApplicationReconciler{ReconcilerBase: controllercommon.NewReconcilerBase(client, nil, scheme, nil, nil)}
	application := &skiperatorv1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{Name: "app", Namespace: "team-a"},
		Spec: skiperatorv1alpha1.ApplicationSpec{
			RoutingProvider: skiperatorv1alpha1.RoutingProviderStandard,
		},
	}

	istioEnabled := reconciler.IsIstioEnabledForNamespace(context.Background(), application.Namespace)
	assert.False(t, istioEnabled)

	err := reconciler.ValidateIstioEnabledForGatewayAPI(application.UsesStandardRouting(), istioEnabled, application.Namespace)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "istio.io/rev")
}

func TestApplicationStandardRoutingAllowsIstioRevision(t *testing.T) {
	scheme := runtime.NewScheme()
	resourceschemas.AddSchemas(scheme)
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "team-a",
			Labels: map[string]string{"istio.io/rev": "istio-1300"},
		},
	}
	client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(namespace).Build()
	reconciler := &ApplicationReconciler{ReconcilerBase: controllercommon.NewReconcilerBase(client, nil, scheme, nil, nil)}
	application := &skiperatorv1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{Name: "app", Namespace: "team-a"},
		Spec: skiperatorv1alpha1.ApplicationSpec{
			RoutingProvider: skiperatorv1alpha1.RoutingProviderStandard,
		},
	}

	istioEnabled := reconciler.IsIstioEnabledForNamespace(context.Background(), application.Namespace)
	assert.True(t, istioEnabled)

	err := reconciler.ValidateIstioEnabledForGatewayAPI(application.UsesStandardRouting(), istioEnabled, application.Namespace)

	require.NoError(t, err)
}
