package dynamic

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
	"k8s.io/apimachinery/pkg/util/intstr"
)

func generatedRoutingIngressRules(t *testing.T, meshMode mesh.Mode) []networkingv1.NetworkPolicyIngressRule {
	t.Helper()

	routing := &skiperatorv1alpha1.Routing{
		ObjectMeta: metav1.ObjectMeta{Name: "api", Namespace: "team-a"},
		Spec: skiperatorv1alpha1.RoutingSpec{
			Hostname: "api.example.com",
			Routes:   []skiperatorv1alpha1.Route{{TargetApp: "backend", PathPrefix: "/", Port: 8080}},
		},
	}
	r := reconciliation.NewRoutingReconciliation(context.Background(), routing, log.NewLogger(), meshMode, nil, nil)

	require.NoError(t, Generate(r))
	require.Len(t, r.GetResources(), 1)

	return r.GetResources()[0].(*networkingv1.NetworkPolicy).Spec.Ingress
}

func TestRoutingAmbientOpensZtunnelPort(t *testing.T) {
	rules := generatedRoutingIngressRules(t, mesh.ModeAmbient)

	require.Len(t, rules, 1)
	// The target app is reached through ztunnel, so the gateway needs the HBONE
	// port as well as the target port.
	assert.Equal(t, []networkingv1.NetworkPolicyPort{
		{Port: new(intstr.FromInt32(8080))},
		{Protocol: new(corev1.ProtocolTCP), Port: new(mesh.ZtunnelInboundPort)},
	}, rules[0].Ports)
}

func TestRoutingSidecarKeepsTargetPortOnly(t *testing.T) {
	rules := generatedRoutingIngressRules(t, mesh.ModeSidecar)

	require.Len(t, rules, 1)
	assert.Equal(t, []networkingv1.NetworkPolicyPort{
		{Port: new(intstr.FromInt32(8080))},
	}, rules[0].Ports)
}
