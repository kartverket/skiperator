package usage

import (
	"context"
	"testing"

	"github.com/kartverket/skiperator/pkg/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
)

func namespaceWithLabels(name string, team string, division string) *corev1.Namespace {
	return &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{
		Name:   name,
		Labels: map[string]string{labelTeam: team, labelDivision: division},
	}}
}

func routable(apiVersion string, kind string, namespace string, name string) *unstructured.Unstructured {
	return &unstructured.Unstructured{Object: map[string]interface{}{
		"apiVersion": apiVersion,
		"kind":       kind,
		"metadata":   map[string]interface{}{"name": name, "namespace": namespace},
	}}
}

// One sweep must cost one List per namespace collection plus one per swept kind,
// no matter how many gauges consume it. A per-namespace fan-out regression shows
// up here as a higher count.
func TestCollectClusterUsageListsEachKindOnce(t *testing.T) {
	scheme := runtime.NewScheme()
	require.NoError(t, corev1.AddToScheme(scheme))

	lists := 0
	c := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(
			namespaceWithLabels("team-a-main", "team-a", "division-1"),
			namespaceWithLabels("team-b-main", "team-b", "division-2"),
			&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "unlabelled"}},
		).
		WithLists(&unstructured.UnstructuredList{
			Object: map[string]interface{}{"apiVersion": "skiperator.kartverket.no/v1alpha1", "kind": "ApplicationList"},
			Items: []unstructured.Unstructured{
				*routable("skiperator.kartverket.no/v1alpha1", "Application", "team-a-main", "app-one"),
				*routable("skiperator.kartverket.no/v1alpha1", "Application", "team-a-main", "app-two"),
				*routable("skiperator.kartverket.no/v1alpha1", "Application", "unlabelled", "app-three"),
			},
		}).
		WithLists(&unstructured.UnstructuredList{
			Object: map[string]interface{}{"apiVersion": "skiperator.kartverket.no/v1alpha1", "kind": "RoutingList"},
			Items: []unstructured.Unstructured{
				*routable("skiperator.kartverket.no/v1alpha1", "Routing", "team-b-main", "routing-one"),
			},
		}).
		WithLists(&unstructured.UnstructuredList{
			Object: map[string]interface{}{"apiVersion": "skiperator.kartverket.no/v1beta1", "kind": "SKIPJobList"},
			Items: []unstructured.Unstructured{
				*routable("skiperator.kartverket.no/v1beta1", "SKIPJob", "team-b-main", "job-one"),
			},
		}).
		WithInterceptorFuncs(interceptor.Funcs{
			List: func(ctx context.Context, c client.WithWatch, list client.ObjectList, opts ...client.ListOption) error {
				lists++
				return c.List(ctx, list, opts...)
			},
		}).
		Build()

	usage, ok := collectClusterUsage(context.Background(), c, log.NewLogger())

	require.True(t, ok)
	assert.Equal(t, 1+len(sweptKinds), lists, "expected one List for namespaces plus one per swept kind")
	assert.Len(t, usage.namespaces, 3)
	require.Len(t, usage.resources, 5)

	// Namespace labels are attached to each resource, and a namespace without
	// them falls back to the unknown value rather than an empty label.
	byName := map[string]usageResource{}
	for _, resource := range usage.resources {
		byName[resource.item.GetName()] = resource
	}
	assert.Equal(t, "team-a", byName["app-one"].team)
	assert.Equal(t, "division-1", byName["app-one"].division)
	assert.Equal(t, typeApplication, byName["app-one"].kind)
	assert.Equal(t, unknownValue, byName["app-three"].team)
	assert.Equal(t, typeSKIPJob, byName["job-one"].kind)

	// SKIPJob has no ingress, so routing gauges must not see it.
	routableKinds := map[string]int{}
	usage.routables(func(resource usageResource) { routableKinds[resource.kind]++ })
	assert.Equal(t, map[string]int{typeApplication: 3, typeRouting: 1}, routableKinds)
}
