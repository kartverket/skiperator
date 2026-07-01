package controllers

import (
	"context"
	"testing"

	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	commontypes "github.com/kartverket/skiperator/api/common"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	controllercommon "github.com/kartverket/skiperator/internal/controllers/common"
	"github.com/kartverket/skiperator/pkg/resourceschemas"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
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

	istioEnabled, err := reconciler.IsIstioEnabledForNamespace(context.Background(), application.Namespace)
	require.NoError(t, err)
	assert.False(t, istioEnabled)

	err = reconciler.ValidateIstioEnabledForGatewayAPI(application.UsesStandardRouting(), istioEnabled, application.Namespace)

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

	istioEnabled, err := reconciler.IsIstioEnabledForNamespace(context.Background(), application.Namespace)
	require.NoError(t, err)
	assert.True(t, istioEnabled)

	err = reconciler.ValidateIstioEnabledForGatewayAPI(application.UsesStandardRouting(), istioEnabled, application.Namespace)

	require.NoError(t, err)
}

func TestRoutingSharedOwnershipSetsSharedResourcesCondition(t *testing.T) {
	routing := &skiperatorv1alpha1.Routing{
		ObjectMeta: metav1.ObjectMeta{Name: "api", Namespace: "team-a"},
		Spec: skiperatorv1alpha1.RoutingSpec{
			Hostname:        "api.example.com",
			RoutingProvider: skiperatorv1alpha1.RoutingProviderStandard,
			Ownership:       skiperatorv1alpha1.RoutingOwnershipShared,
			Routes:          []skiperatorv1alpha1.Route{{TargetApp: "backend", PathPrefix: "/", Port: 8080}},
		},
	}

	setSharedRoutingResourcesCondition(routing)

	condition := meta.FindStatusCondition(routing.Status.Conditions, commontypes.SharedRoutingResourcesType)
	if assert.NotNil(t, condition) {
		assert.Equal(t, metav1.ConditionTrue, condition.Status)
		assert.Equal(t, "SharedRoutingResourcesActive", condition.Reason)
	}
}

func TestRoutingCertificateWatchUsesRoutingLabels(t *testing.T) {
	reconciler := &RoutingReconciler{}
	certificate := &certmanagerv1.Certificate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cert",
			Namespace: "team-a",
			Labels: map[string]string{
				"app.kubernetes.io/managed-by":              "skiperator",
				"skiperator.kartverket.no/controller":       "routing",
				"skiperator.kartverket.no/source-namespace": "team-a",
				"skiperator.kartverket.no/routing-name":     "api",
				"application.skiperator.no/app-namespace":   "wrong-namespace",
				"application.skiperator.no/app-name":        "wrong-name",
			},
		},
	}

	requests := reconciler.skiperatorRoutingCertRequests(context.Background(), certificate)

	require.Len(t, requests, 1)
	assert.Equal(t, "team-a", requests[0].Namespace)
	assert.Equal(t, "api", requests[0].Name)
}
