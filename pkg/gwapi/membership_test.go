package gwapi

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
)

func membershipClient(t *testing.T) *fake.ClientBuilder {
	t.Helper()
	s := runtime.NewScheme()
	require.NoError(t, scheme.AddToScheme(s))
	return fake.NewClientBuilder().WithScheme(s)
}

func TestRegisterSharedContributorCreatesAndAccumulates(t *testing.T) {
	ctx := context.Background()
	c := membershipClient(t).Build()
	hostname := "shared.example.com"

	require.NoError(t, RegisterSharedContributor(ctx, c, hostname, types.NamespacedName{Namespace: "team-a", Name: "api"}))
	// Idempotent re-register and a second contributor.
	require.NoError(t, RegisterSharedContributor(ctx, c, hostname, types.NamespacedName{Namespace: "team-a", Name: "api"}))
	require.NoError(t, RegisterSharedContributor(ctx, c, hostname, types.NamespacedName{Namespace: "team-b", Name: "web"}))

	cm := &corev1.ConfigMap{}
	require.NoError(t, c.Get(ctx, types.NamespacedName{Namespace: IstioGatewayNamespace, Name: SharedMembershipName(hostname)}, cm))
	assert.Len(t, cm.Data, 2)
	assert.Contains(t, cm.Data, "team-a.api")
	assert.Contains(t, cm.Data, "team-b.web")
}

func TestDeregisterSharedContributorReportsEmptyOnlyWhenLast(t *testing.T) {
	ctx := context.Background()
	c := membershipClient(t).Build()
	hostname := "shared.example.com"
	require.NoError(t, RegisterSharedContributor(ctx, c, hostname, types.NamespacedName{Namespace: "team-a", Name: "api"}))
	require.NoError(t, RegisterSharedContributor(ctx, c, hostname, types.NamespacedName{Namespace: "team-b", Name: "web"}))

	empty, err := DeregisterSharedContributor(ctx, c, hostname, types.NamespacedName{Namespace: "team-a", Name: "api"})
	require.NoError(t, err)
	assert.False(t, empty)

	// CM still exists with one contributor.
	cm := &corev1.ConfigMap{}
	require.NoError(t, c.Get(ctx, types.NamespacedName{Namespace: IstioGatewayNamespace, Name: SharedMembershipName(hostname)}, cm))

	empty, err = DeregisterSharedContributor(ctx, c, hostname, types.NamespacedName{Namespace: "team-b", Name: "web"})
	require.NoError(t, err)
	assert.True(t, empty)

	// CM must be deleted once the last contributor leaves.
	err = c.Get(ctx, types.NamespacedName{Namespace: IstioGatewayNamespace, Name: SharedMembershipName(hostname)}, &corev1.ConfigMap{})
	assert.True(t, errors.IsNotFound(err), "ConfigMap must be deleted after last contributor deregisters")
}

// TestDeregisterSharedContributorConcurrentRegister verifies the TOCTOU fix:
// if a concurrent RegisterSharedContributor adds a key between our read and
// delete, the Delete returns Conflict, RetryOnConflict re-reads the CM, and
// Deregister correctly returns empty=false without deleting the CM.
func TestDeregisterSharedContributorConcurrentRegister(t *testing.T) {
	ctx := context.Background()
	hostname := "shared.example.com"
	contributor := types.NamespacedName{Namespace: "team-a", Name: "api"}
	newcomer := types.NamespacedName{Namespace: "team-c", Name: "svc"}

	// Seed CM with just contributor so Deregister will try to delete it.
	existing := newMembershipConfigMap(hostname)
	existing.Data = map[string]string{contributorKey(contributor): ""}

	// On the first Delete, inject a concurrent Register (add newcomer to the store),
	// then return Conflict so RetryOnConflict re-reads the updated CM.
	deleteCalled := false
	c := membershipClient(t).
		WithObjects(existing).
		WithInterceptorFuncs(interceptor.Funcs{
			Delete: func(ctx context.Context, cl client.WithWatch, obj client.Object, opts ...client.DeleteOption) error {
				if _, ok := obj.(*corev1.ConfigMap); ok && !deleteCalled {
					deleteCalled = true
					key := types.NamespacedName{Namespace: obj.GetNamespace(), Name: obj.GetName()}
					cm := &corev1.ConfigMap{}
					if err := cl.Get(ctx, key, cm); err != nil {
						return err
					}
					if cm.Data == nil {
						cm.Data = map[string]string{}
					}
					cm.Data[contributorKey(newcomer)] = ""
					if err := cl.Update(ctx, cm); err != nil {
						return err
					}
					return errors.NewConflict(corev1.Resource("configmaps"), obj.GetName(), nil)
				}
				return cl.Delete(ctx, obj, opts...)
			},
		}).
		Build()

	empty, err := DeregisterSharedContributor(ctx, c, hostname, contributor)
	require.NoError(t, err)
	assert.False(t, empty, "concurrent register must prevent empty=true")
	assert.True(t, deleteCalled, "delete must have been attempted")

	// CM must still exist with only the newcomer.
	cm := &corev1.ConfigMap{}
	require.NoError(t, c.Get(ctx, types.NamespacedName{Namespace: IstioGatewayNamespace, Name: SharedMembershipName(hostname)}, cm))
	assert.NotContains(t, cm.Data, contributorKey(contributor))
	assert.Contains(t, cm.Data, contributorKey(newcomer))
}

func TestDeregisterSharedContributorMissingConfigMapIsEmpty(t *testing.T) {
	ctx := context.Background()
	c := membershipClient(t).Build()

	empty, err := DeregisterSharedContributor(ctx, c, "shared.example.com", types.NamespacedName{Namespace: "team-a", Name: "api"})
	require.NoError(t, err)
	assert.True(t, empty)
}

func TestRegisterSharedContributorHandlesCreateRace(t *testing.T) {
	ctx := context.Background()
	hostname := "shared.example.com"
	contributor := types.NamespacedName{Namespace: "team-a", Name: "api"}

	// Pre-seed ConfigMap as if another writer already created it (without our contributor key).
	existing := newMembershipConfigMap(hostname)
	existing.Data = map[string]string{"team-b.web": ""}

	// Simulate the create race:
	//   1. Intercept the first Get to return NotFound (our goroutine reads before the CM exists).
	//   2. We then call Create; the fake client returns AlreadyExists because CM is in the store.
	//   3. RegisterSharedContributor re-reads and falls through to Update.
	getCount := 0
	c := membershipClient(t).
		WithObjects(existing).
		WithInterceptorFuncs(interceptor.Funcs{
			Get: func(ctx context.Context, cl client.WithWatch, key types.NamespacedName, obj client.Object, opts ...client.GetOption) error {
				if _, ok := obj.(*corev1.ConfigMap); ok && getCount == 0 {
					getCount++
					return errors.NewNotFound(corev1.Resource("configmaps"), key.Name)
				}
				getCount++
				return cl.Get(ctx, key, obj, opts...)
			},
		}).
		Build()

	require.NoError(t, RegisterSharedContributor(ctx, c, hostname, contributor))
	assert.GreaterOrEqual(t, getCount, 2, "must re-read after AlreadyExists")

	cm := &corev1.ConfigMap{}
	require.NoError(t, c.Get(ctx, types.NamespacedName{Namespace: IstioGatewayNamespace, Name: SharedMembershipName(hostname)}, cm))
	assert.Contains(t, cm.Data, "team-a.api")
	assert.Contains(t, cm.Data, "team-b.web")
}

func TestDeleteSharedMembershipIsIdempotent(t *testing.T) {
	ctx := context.Background()
	c := membershipClient(t).Build()
	hostname := "shared.example.com"
	require.NoError(t, RegisterSharedContributor(ctx, c, hostname, types.NamespacedName{Namespace: "team-a", Name: "api"}))

	require.NoError(t, DeleteSharedMembership(ctx, c, hostname))
	// Second delete on a missing ConfigMap must not error.
	require.NoError(t, DeleteSharedMembership(ctx, c, hostname))

	err := c.Get(ctx, types.NamespacedName{Namespace: IstioGatewayNamespace, Name: SharedMembershipName(hostname)}, &corev1.ConfigMap{})
	assert.True(t, errors.IsNotFound(err))
}
