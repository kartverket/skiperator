package dynamic

import (
	"context"
	"testing"

	"github.com/kartverket/skiperator/api/common/podtypes"
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

func TestGetInboundIngressRules(t *testing.T) {
	tcpPort := func(port int32) networkingv1.NetworkPolicyPort {
		return networkingv1.NetworkPolicyPort{Protocol: new(corev1.ProtocolTCP), Port: new(intstr.FromInt32(port))}
	}
	peer := func(app string) networkingv1.NetworkPolicyPeer {
		return networkingv1.NetworkPolicyPeer{
			NamespaceSelector: &metav1.LabelSelector{MatchLabels: map[string]string{"kubernetes.io/metadata.name": "team-a"}},
			PodSelector:       &metav1.LabelSelector{MatchLabels: map[string]string{"app": app}},
		}
	}
	appPort := networkingv1.NetworkPolicyPort{Port: new(intstr.FromInt32(8080))}
	ztunnelPort := networkingv1.NetworkPolicyPort{Protocol: new(corev1.ProtocolTCP), Port: new(mesh.ZtunnelInboundPort)}

	tests := []struct {
		name     string
		rules    []podtypes.InternalRule
		port     int32
		meshMode mesh.Mode
		want     []networkingv1.NetworkPolicyIngressRule
	}{
		{
			name:     "rules without ports share one rule for the inbound port",
			rules:    []podtypes.InternalRule{{Application: "a"}, {Application: "b"}},
			port:     8080,
			meshMode: mesh.ModeSidecar,
			want: []networkingv1.NetworkPolicyIngressRule{
				{From: []networkingv1.NetworkPolicyPeer{peer("a"), peer("b")}, Ports: []networkingv1.NetworkPolicyPort{appPort}},
			},
		},
		{
			name:     "rule with ports opens them in addition to the inbound port",
			rules:    []podtypes.InternalRule{{Application: "keycloak", Ports: []networkingv1.NetworkPolicyPort{tcpPort(57800), tcpPort(7800)}}},
			port:     8080,
			meshMode: mesh.ModeSidecar,
			want: []networkingv1.NetworkPolicyIngressRule{
				{From: []networkingv1.NetworkPolicyPeer{peer("keycloak")}, Ports: []networkingv1.NetworkPolicyPort{appPort, tcpPort(57800), tcpPort(7800)}},
			},
		},
		{
			name:     "ambient keeps the ztunnel port for a rule with ports",
			rules:    []podtypes.InternalRule{{Application: "keycloak", Ports: []networkingv1.NetworkPolicyPort{tcpPort(7800)}}},
			port:     8080,
			meshMode: mesh.ModeAmbient,
			want: []networkingv1.NetworkPolicyIngressRule{
				{From: []networkingv1.NetworkPolicyPeer{peer("keycloak")}, Ports: []networkingv1.NetworkPolicyPort{appPort, ztunnelPort, tcpPort(7800)}},
			},
		},
		{
			name: "rules with and without ports get separate rules",
			rules: []podtypes.InternalRule{
				{Application: "a"},
				{Application: "keycloak", Ports: []networkingv1.NetworkPolicyPort{tcpPort(7800)}},
			},
			port:     8080,
			meshMode: mesh.ModeSidecar,
			want: []networkingv1.NetworkPolicyIngressRule{
				{From: []networkingv1.NetworkPolicyPeer{peer("a")}, Ports: []networkingv1.NetworkPolicyPort{appPort}},
				{From: []networkingv1.NetworkPolicyPeer{peer("keycloak")}, Ports: []networkingv1.NetworkPolicyPort{appPort, tcpPort(7800)}},
			},
		},
		{
			name:     "without an inbound port only the ports of the rule are opened",
			rules:    []podtypes.InternalRule{{Application: "a", Ports: []networkingv1.NetworkPolicyPort{tcpPort(7800)}}},
			meshMode: mesh.ModeSidecar,
			want: []networkingv1.NetworkPolicyIngressRule{
				{From: []networkingv1.NetworkPolicyPeer{peer("a")}, Ports: []networkingv1.NetworkPolicyPort{tcpPort(7800)}},
			},
		},
		{
			name:     "empty inbound rules keep one rule without sources",
			port:     8080,
			meshMode: mesh.ModeSidecar,
			want: []networkingv1.NetworkPolicyIngressRule{
				{Ports: []networkingv1.NetworkPolicyPort{appPort}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getInboundIngressRules(tt.rules, "team-a", tt.port, tt.meshMode)

			assert.Equal(t, tt.want, got)
		})
	}
}
