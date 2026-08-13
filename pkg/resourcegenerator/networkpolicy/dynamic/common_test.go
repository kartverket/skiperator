package dynamic

import (
	"context"
	"testing"

	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/internal/config"
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

func generatedIngressRules(t *testing.T, meshMode mesh.Mode) []networkingv1.NetworkPolicyIngressRule {
	t.Helper()

	application := &skiperatorv1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{Name: "app", Namespace: "team-a"},
		Spec: skiperatorv1alpha1.ApplicationSpec{
			Image:         "image",
			Port:          8080,
			Ingresses:     []string{"app.example.com"},
			IstioSettings: &skiperatorv1alpha1.IstioSettingsApplication{},
		},
	}
	application.FillDefaultsSpec()
	r := reconciliation.NewApplicationReconciliation(context.Background(), application, log.NewLogger(), meshMode, nil, nil, config.SkiperatorConfig{})

	require.NoError(t, Generate(r))
	require.Len(t, r.GetResources(), 1)

	return r.GetResources()[0].(*networkingv1.NetworkPolicy).Spec.Ingress
}

func TestAmbientOpensZtunnelPortForGatewayTraffic(t *testing.T) {
	rules := generatedIngressRules(t, mesh.ModeAmbient)

	require.NotEmpty(t, rules)
	// The gateway keeps its source restriction, but ambient delivers the
	// request through ztunnel's HBONE port instead of the application port.
	assert.Equal(t, []networkingv1.NetworkPolicyPort{
		{Port: new(intstr.FromInt32(8080))},
		{Protocol: new(corev1.ProtocolTCP), Port: new(mesh.ZtunnelInboundPort)},
	}, rules[0].Ports)
}

func TestSidecarKeepsApplicationPortOnly(t *testing.T) {
	rules := generatedIngressRules(t, mesh.ModeSidecar)

	require.NotEmpty(t, rules)
	assert.Equal(t, []networkingv1.NetworkPolicyPort{
		{Port: new(intstr.FromInt32(8080))},
	}, rules[0].Ports)
}
