package controllers

import (
	"context"
	"testing"

	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	controllercommon "github.com/kartverket/skiperator/internal/controllers/common"
	"github.com/kartverket/skiperator/pkg/gwapi"
	"github.com/kartverket/skiperator/pkg/log"
	"github.com/kartverket/skiperator/pkg/resourceschemas"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	gatewayapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

func sharedRouting(namespace, name, hostname string) *skiperatorv1alpha1.Routing {
	return &skiperatorv1alpha1.Routing{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
		Spec: skiperatorv1alpha1.RoutingSpec{
			Hostname:        hostname,
			RoutingProvider: skiperatorv1alpha1.RoutingProviderStandard,
			Ownership:       skiperatorv1alpha1.RoutingOwnershipShared,
		},
	}
}

func TestSharedRoutingDeletesSharedResourcesOnlyForLastContributor(t *testing.T) {
	ctx := context.Background()
	hostname := "shared.example.com"
	scheme := runtime.NewScheme()
	resourceschemas.AddSchemas(scheme)

	teamA := sharedRouting("team-a", "api", hostname)
	teamB := sharedRouting("team-b", "web", hostname)
	sharedListenerSet := &gatewayapiv1.ListenerSet{
		ObjectMeta: metav1.ObjectMeta{Namespace: gwapi.IstioGatewayNamespace, Name: gwapi.SharedListenerSetName(hostname)},
	}
	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(teamA, teamB, sharedListenerSet).Build()
	reconciler := &RoutingReconciler{ReconcilerBase: controllercommon.NewReconcilerBase(c, nil, scheme, nil, nil)}

	// Both contributors register their membership.
	require.NoError(t, gwapi.RegisterSharedContributor(ctx, c, hostname, types.NamespacedName{Namespace: "team-a", Name: "api"}))
	require.NoError(t, gwapi.RegisterSharedContributor(ctx, c, hostname, types.NamespacedName{Namespace: "team-b", Name: "web"}))

	// Releasing team-a leaves team-b: shared resources are kept.
	require.NoError(t, reconciler.releaseSharedMembership(ctx, teamA, log.NewLogger()))
	err := c.Get(ctx, types.NamespacedName{Namespace: gwapi.IstioGatewayNamespace, Name: gwapi.SharedListenerSetName(hostname)}, &gatewayapiv1.ListenerSet{})
	require.NoError(t, err)

	// Releasing the last contributor deletes the shared resources and membership.
	require.NoError(t, reconciler.releaseSharedMembership(ctx, teamB, log.NewLogger()))
	err = c.Get(ctx, types.NamespacedName{Namespace: gwapi.IstioGatewayNamespace, Name: gwapi.SharedListenerSetName(hostname)}, &gatewayapiv1.ListenerSet{})
	assert.True(t, apierrors.IsNotFound(err))
	err = c.Get(ctx, types.NamespacedName{Namespace: gwapi.IstioGatewayNamespace, Name: gwapi.SharedMembershipName(hostname)}, &corev1.ConfigMap{})
	assert.True(t, apierrors.IsNotFound(err))
}

func TestSharedRoutingRegistersMembershipBeforeApplyingResources(t *testing.T) {
	ctx := context.Background()
	hostname := "shared.example.com"
	scheme := runtime.NewScheme()
	resourceschemas.AddSchemas(scheme)

	// The fixture is the smallest valid shared Gateway API Routing reconcile:
	// an Istio-enabled namespace, one target Application for route port lookup,
	// and one shared Routing that should create shared istio-gateways resources.
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "team-a",
			Labels: map[string]string{"istio.io/rev": "revision"},
		},
	}
	application := &skiperatorv1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{Name: "app", Namespace: "team-a"},
		Spec: skiperatorv1alpha1.ApplicationSpec{
			Image: "image",
			Port:  8080,
		},
	}
	routing := sharedRouting("team-a", "api", hostname)
	routing.Spec.Routes = []skiperatorv1alpha1.Route{{TargetApp: "app", PathPrefix: "/", Port: 8080}}
	c := fake.NewClientBuilder().
		WithScheme(scheme).
		WithStatusSubresource(routing).
		WithObjects(namespace, application, routing).
		Build()
	reconciler := &RoutingReconciler{
		ReconcilerBase: controllercommon.NewReconcilerBase(c, nil, scheme, nil, record.NewFakeRecorder(10)),
	}

	// First pass adds the finalizer and stops. Deletion cleanup depends on this finalizer, but membership registration happens in a later full reconcile.
	_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: types.NamespacedName{Namespace: "team-a", Name: "api"}})
	require.NoError(t, err)
	updatedRouting := &skiperatorv1alpha1.Routing{}
	require.NoError(t, c.Get(ctx, types.NamespacedName{Namespace: "team-a", Name: "api"}, updatedRouting))
	require.True(t, ctrlutil.ContainsFinalizer(updatedRouting, sharedRoutingFinalizer))

	// Drive reconcile until the shared ListenerSet has actually been applied.
	req := reconcile.Request{NamespacedName: types.NamespacedName{Namespace: "team-a", Name: "api"}}
	for i := 0; i < 10; i++ {
		_, err = reconciler.Reconcile(ctx, req)
		require.NoError(t, err)
		err = c.Get(ctx, types.NamespacedName{Namespace: gwapi.IstioGatewayNamespace, Name: gwapi.SharedListenerSetName(hostname)}, &gatewayapiv1.ListenerSet{})
		if err == nil {
			break
		}
		require.True(t, apierrors.IsNotFound(err))
	}

	// Once shared resources are visible, the contributor must already be in the
	// membership ConfigMap. Otherwise deleting another contributor can see an
	// incomplete membership set and delete shared resources too early.
	cm := &corev1.ConfigMap{}
	require.NoError(t, c.Get(ctx, types.NamespacedName{Namespace: gwapi.IstioGatewayNamespace, Name: gwapi.SharedMembershipName(hostname)}, cm))
	assert.Contains(t, cm.Data, "team-a.api")
	require.NoError(t, c.Get(ctx, types.NamespacedName{Namespace: gwapi.IstioGatewayNamespace, Name: gwapi.SharedListenerSetName(hostname)}, &gatewayapiv1.ListenerSet{}))
}

// A custom certificate is provisioned manually into istio-gateways and is not
// owned by skiperator, so the GC must never delete it (only the shared
// ListenerSet, which skiperator does own).
func TestSharedRoutingDoesNotDeleteCustomCertificate(t *testing.T) {
	ctx := context.Background()
	scheme := runtime.NewScheme()
	resourceschemas.AddSchemas(scheme)

	routing := sharedRouting("team-a", "api", "shared.example.com+manual-tls")
	host, err := routing.Spec.GetHost()
	require.NoError(t, err)
	require.True(t, host.UsesCustomCert())

	customCert := &certmanagerv1.Certificate{
		ObjectMeta: metav1.ObjectMeta{Namespace: gwapi.IstioGatewayNamespace, Name: "manual-tls"},
	}
	sharedListenerSet := &gatewayapiv1.ListenerSet{
		ObjectMeta: metav1.ObjectMeta{Namespace: gwapi.IstioGatewayNamespace, Name: gwapi.SharedListenerSetName(host.Hostname)},
	}
	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(routing, customCert, sharedListenerSet).Build()
	reconciler := &RoutingReconciler{ReconcilerBase: controllercommon.NewReconcilerBase(c, nil, scheme, nil, nil)}

	require.NoError(t, gwapi.RegisterSharedContributor(ctx, c, host.Hostname, types.NamespacedName{Namespace: "team-a", Name: "api"}))
	require.NoError(t, reconciler.releaseSharedMembership(ctx, routing, log.NewLogger()))

	// The shared ListenerSet (skiperator-owned) is removed.
	err = c.Get(ctx, types.NamespacedName{Namespace: gwapi.IstioGatewayNamespace, Name: gwapi.SharedListenerSetName(host.Hostname)}, &gatewayapiv1.ListenerSet{})
	assert.True(t, apierrors.IsNotFound(err))
	// The manually-provisioned custom certificate is left untouched.
	err = c.Get(ctx, types.NamespacedName{Namespace: gwapi.IstioGatewayNamespace, Name: "manual-tls"}, &certmanagerv1.Certificate{})
	require.NoError(t, err)
}
