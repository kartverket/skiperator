package virtualservice

import (
	"context"
	"testing"

	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/internal/config"
	"github.com/kartverket/skiperator/pkg/log"
	"github.com/kartverket/skiperator/pkg/reconciliation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	istionetworkingv1 "istio.io/client-go/pkg/apis/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestApplicationVirtualServiceSkipsLegacyWhenDisabled(t *testing.T) {
	application := &skiperatorv1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{Name: "app", Namespace: "team-a"},
		Spec: skiperatorv1alpha1.ApplicationSpec{
			Image:           "image",
			Port:            8080,
			Ingresses:       []string{"app.example.com"},
			RoutingProvider: skiperatorv1alpha1.RoutingProviderStandard,
		},
	}
	r := reconciliation.NewApplicationReconciliation(context.Background(), application, log.NewLogger(), false, nil, nil, config.SkiperatorConfig{})
	r.SetGenerateLegacyRouting(false)

	err := Generate(r)

	require.NoError(t, err)
	require.Empty(t, r.GetResources())
}

func TestApplicationLegacyRoutingGeneratesVirtualService(t *testing.T) {
	application := &skiperatorv1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{Name: "app", Namespace: "team-a"},
		Spec: skiperatorv1alpha1.ApplicationSpec{
			Image:           "image",
			Port:            8080,
			Ingresses:       []string{"app.example.com"},
			RoutingProvider: skiperatorv1alpha1.RoutingProviderLegacy,
			IstioSettings:   &skiperatorv1alpha1.IstioSettingsApplication{},
		},
	}
	r := reconciliation.NewApplicationReconciliation(context.Background(), application, log.NewLogger(), false, nil, nil, config.SkiperatorConfig{})

	err := Generate(r)

	require.NoError(t, err)
	require.Len(t, r.GetResources(), 1)
	virtualService := r.GetResources()[0].(*istionetworkingv1.VirtualService)
	assert.Equal(t, "team-a", virtualService.Namespace)
	assert.Equal(t, application.GetVirtualServiceName(), virtualService.Name)
	assert.Equal(t, []string{"app.example.com"}, virtualService.Spec.Hosts)
}

func TestRoutingLegacyRoutingGeneratesVirtualService(t *testing.T) {
	routing := &skiperatorv1alpha1.Routing{
		ObjectMeta: metav1.ObjectMeta{Name: "api", Namespace: "team-a"},
		Spec: skiperatorv1alpha1.RoutingSpec{
			Hostname:        "api.example.com",
			RoutingProvider: skiperatorv1alpha1.RoutingProviderLegacy,
			Ownership:       skiperatorv1alpha1.RoutingOwnershipShared,
			Routes:          []skiperatorv1alpha1.Route{{TargetApp: "backend", PathPrefix: "/", Port: 8080}},
		},
	}
	r := reconciliation.NewRoutingReconciliation(context.Background(), routing, log.NewLogger(), false, nil)

	err := Generate(r)

	require.NoError(t, err)
	require.Len(t, r.GetResources(), 1)
	virtualService := r.GetResources()[0].(*istionetworkingv1.VirtualService)
	assert.Equal(t, "team-a", virtualService.Namespace)
	assert.Equal(t, routing.GetVirtualServiceName(), virtualService.Name)
	assert.Equal(t, []string{"api.example.com"}, virtualService.Spec.Hosts)
}
