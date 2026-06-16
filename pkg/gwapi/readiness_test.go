package gwapi

import (
	"context"
	"fmt"
	"testing"

	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	istionetworkingv1 "istio.io/client-go/pkg/apis/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

type transientGetClient struct {
	client.Client
	failures int
	gets     int
}

func (c *transientGetClient) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
	if c.gets < c.failures {
		c.gets++
		return errors.NewInternalError(fmt.Errorf("transient"))
	}
	c.gets++
	return c.Client.Get(ctx, key, obj, opts...)
}

func TestLegacyRoutingExistsRetriesTransientGetFailure(t *testing.T) {
	scheme := runtime.NewScheme()
	require.NoError(t, istionetworkingv1.AddToScheme(scheme))
	routing := &skiperatorv1alpha1.Routing{ObjectMeta: metav1.ObjectMeta{Name: "api", Namespace: "team-a"}}
	virtualService := &istionetworkingv1.VirtualService{ObjectMeta: metav1.ObjectMeta{Name: "api-routing-ingress", Namespace: "team-a"}}
	c := &transientGetClient{
		Client:   fake.NewClientBuilder().WithScheme(scheme).WithObjects(virtualService).Build(),
		failures: 1,
	}

	exists, err := legacyRoutingExists(context.Background(), c, routing)
	require.NoError(t, err)
	assert.True(t, exists)
	assert.GreaterOrEqual(t, c.gets, 2)
}

func TestLegacyRoutingExistsReturnsErrorOnPersistentFailure(t *testing.T) {
	scheme := runtime.NewScheme()
	require.NoError(t, istionetworkingv1.AddToScheme(scheme))
	routing := &skiperatorv1alpha1.Routing{ObjectMeta: metav1.ObjectMeta{Name: "api", Namespace: "team-a"}}
	c := &transientGetClient{
		Client:   fake.NewClientBuilder().WithScheme(scheme).Build(),
		failures: 1000,
	}

	// A persistent non-NotFound error must surface as an error, not be read as
	// "legacy absent" (which would prune legacy routing mid-migration).
	exists, err := legacyRoutingExists(context.Background(), c, routing)
	require.Error(t, err)
	assert.False(t, exists)
}

func TestLegacyRoutingExistsDoesNotRetryNotFound(t *testing.T) {
	scheme := runtime.NewScheme()
	require.NoError(t, istionetworkingv1.AddToScheme(scheme))
	routing := &skiperatorv1alpha1.Routing{ObjectMeta: metav1.ObjectMeta{Name: "api", Namespace: "team-a"}}
	c := &transientGetClient{
		Client: fake.NewClientBuilder().WithScheme(scheme).Build(),
	}

	exists, err := legacyRoutingExists(context.Background(), c, routing)
	require.NoError(t, err)
	assert.False(t, exists)
	assert.Equal(t, 2, c.gets)
}
