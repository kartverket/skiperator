package controllers

import (
	"context"
	"testing"

	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	controllercommon "github.com/kartverket/skiperator/internal/controllers/common"
	"github.com/kartverket/skiperator/pkg/gwapi"
	"github.com/kartverket/skiperator/pkg/resourceschemas"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	// Two live contributors: team-a is not the last.
	last, err := reconciler.isLastSharedContributor(ctx, teamA, hostname)
	require.NoError(t, err)
	assert.False(t, last)

	// Once the other contributor is gone, team-a is the last.
	require.NoError(t, c.Delete(ctx, teamB))
	last, err = reconciler.isLastSharedContributor(ctx, teamA, hostname)
	require.NoError(t, err)
	assert.True(t, last)

	// Deleting shared resources removes the shared ListenerSet, and is idempotent.
	host, err := teamA.Spec.GetHost()
	require.NoError(t, err)
	require.NoError(t, reconciler.deleteSharedRoutingResources(ctx, teamA, host))
	err = c.Get(ctx, types.NamespacedName{Namespace: gwapi.IstioGatewayNamespace, Name: gwapi.SharedListenerSetName(hostname)}, &gatewayapiv1.ListenerSet{})
	assert.True(t, apierrors.IsNotFound(err))
	require.NoError(t, reconciler.deleteSharedRoutingResources(ctx, teamA, host))
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

	require.NoError(t, reconciler.deleteSharedRoutingResources(ctx, routing, host))

	// The shared ListenerSet (skiperator-owned) is removed.
	err = c.Get(ctx, types.NamespacedName{Namespace: gwapi.IstioGatewayNamespace, Name: gwapi.SharedListenerSetName(host.Hostname)}, &gatewayapiv1.ListenerSet{})
	assert.True(t, apierrors.IsNotFound(err))
	// The manually-provisioned custom certificate is left untouched.
	err = c.Get(ctx, types.NamespacedName{Namespace: gwapi.IstioGatewayNamespace, Name: "manual-tls"}, &certmanagerv1.Certificate{})
	require.NoError(t, err)
}

// A contributor that is itself being deleted must not keep shared resources
// alive for the genuinely-last contributor.
func TestSharedRoutingIgnoresDeletingContributors(t *testing.T) {
	ctx := context.Background()
	hostname := "shared.example.com"
	scheme := runtime.NewScheme()
	resourceschemas.AddSchemas(scheme)

	teamA := sharedRouting("team-a", "api", hostname)
	teamB := sharedRouting("team-b", "web", hostname)
	deletionTime := metav1.Now()
	teamB.DeletionTimestamp = &deletionTime
	teamB.Finalizers = []string{sharedRoutingFinalizer}

	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(teamA, teamB).Build()
	reconciler := &RoutingReconciler{ReconcilerBase: controllercommon.NewReconcilerBase(c, nil, scheme, nil, nil)}

	// team-b is being deleted, so team-a is effectively the last live contributor.
	last, err := reconciler.isLastSharedContributor(ctx, teamA, hostname)
	require.NoError(t, err)
	assert.True(t, last)
}
