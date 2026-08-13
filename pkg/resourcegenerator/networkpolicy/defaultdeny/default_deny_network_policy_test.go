package defaultdeny

import (
	"context"
	"testing"

	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/log"
	"github.com/kartverket/skiperator/pkg/mesh"
	"github.com/kartverket/skiperator/pkg/reconciliation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func generatedDefaultDeny(t *testing.T, meshMode mesh.Mode) *networkingv1.NetworkPolicy {
	t.Helper()

	namespace := skiperatorv1alpha1.SKIPNamespace{
		Namespace: &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "team-a"}},
	}
	r := reconciliation.NewNamespaceReconciliation(context.Background(), namespace, log.NewLogger(), meshMode, nil)

	generator, err := NewDefaultDenyNetworkPolicy(nil, false)
	require.NoError(t, err)
	require.NoError(t, generator.Generate(r))
	require.Len(t, r.GetResources(), 1)

	return r.GetResources()[0].(*networkingv1.NetworkPolicy)
}

func TestAmbientAllowsKubeletHealthProbes(t *testing.T) {
	// Ambient SNATs probes to a link-local address, which the deny all ingress
	// rule would otherwise drop for every pod in the namespace.
	ingress := generatedDefaultDeny(t, mesh.ModeAmbient).Spec.Ingress

	require.Len(t, ingress, 1)
	require.Len(t, ingress[0].From, 1)
	assert.Equal(t, mesh.AmbientHealthProbeCIDR, ingress[0].From[0].IPBlock.CIDR)
}

func TestSidecarKeepsDenyingAllIngress(t *testing.T) {
	assert.Empty(t, generatedDefaultDeny(t, mesh.ModeSidecar).Spec.Ingress)
}
