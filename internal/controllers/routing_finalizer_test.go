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
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
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
