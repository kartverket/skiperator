package gatewayapi

import (
	"context"
	"testing"

	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/internal/config"
	"github.com/kartverket/skiperator/pkg/gwapi"
	"github.com/kartverket/skiperator/pkg/log"
	"github.com/kartverket/skiperator/pkg/reconciliation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gatewayapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

func TestApplicationStandardRoutingWithoutIngresses(t *testing.T) {
	app := &skiperatorv1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{Name: "app", Namespace: "team-a"},
		Spec: skiperatorv1alpha1.ApplicationSpec{
			Image:           "image",
			Port:            8080,
			RoutingProvider: skiperatorv1alpha1.RoutingProviderStandard,
			RedirectToHTTPS: skiperatorv1alpha1Bool(true),
		},
	}
	r := reconciliation.NewApplicationReconciliation(context.Background(), app, log.NewLogger(), false, nil, nil, config.SkiperatorConfig{})

	require.NoError(t, Generate(r))

	// Routes without parentRefs would never be programmed, so none are created.
	assert.Empty(t, r.GetResources())
}

func TestApplicationStandardRouting(t *testing.T) {
	app := &skiperatorv1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{Name: "app", Namespace: "team-a"},
		Spec: skiperatorv1alpha1.ApplicationSpec{
			Image:           "image",
			Port:            8080,
			Ingresses:       []string{"app.example.com"},
			RoutingProvider: skiperatorv1alpha1.RoutingProviderStandard,
			RedirectToHTTPS: skiperatorv1alpha1Bool(true),
		},
	}
	r := reconciliation.NewApplicationReconciliation(context.Background(), app, log.NewLogger(), false, nil, nil, config.SkiperatorConfig{})

	err := Generate(r)

	require.NoError(t, err)
	require.Len(t, r.GetResources(), 3)

	listenerSet := r.GetResources()[0].(*gatewayapiv1.ListenerSet)
	assert.Equal(t, "team-a", listenerSet.Namespace)
	assert.Equal(t, gatewayapiv1.ObjectName(gwapi.ExternalGatewayName), listenerSet.Spec.ParentRef.Name)
	assert.Equal(t, "app.example.com", string(*listenerSet.Spec.Listeners[1].Hostname))
	assert.Equal(t, gatewayapiv1.ObjectName("team-a-app-ingress-7f92f5cfd8862fd3"), listenerSet.Spec.Listeners[1].TLS.CertificateRefs[0].Name)
	// Namespace-local listeners keep the default route scope (no AllowedRoutes).
	assert.Nil(t, listenerSet.Spec.Listeners[1].AllowedRoutes)

	redirectRoute := r.GetResources()[1].(*gatewayapiv1.HTTPRoute)
	assert.Equal(t, "app-redirect", redirectRoute.Name)
	assert.Equal(t, httpSectionName, *redirectRoute.Spec.ParentRefs[0].SectionName)
	assert.Equal(t, gatewayapiv1.HTTPRouteFilterRequestRedirect, redirectRoute.Spec.Rules[0].Filters[0].Type)

	route := r.GetResources()[2].(*gatewayapiv1.HTTPRoute)
	assert.Equal(t, "team-a", route.Namespace)
	assert.Equal(t, gatewayapiv1.Kind("ListenerSet"), *route.Spec.ParentRefs[0].Kind)
	assert.Equal(t, httpsSectionName, *route.Spec.ParentRefs[0].SectionName)
	assert.Equal(t, gatewayapiv1.ObjectName("app"), route.Spec.Rules[0].BackendRefs[0].Name)
}

func TestApplicationStandardRoutingWithExtraContainerUsesServicePort(t *testing.T) {
	proxyPort := int32(8443)
	app := &skiperatorv1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{Name: "app", Namespace: "team-a"},
		Spec: skiperatorv1alpha1.ApplicationSpec{
			Image:           "image",
			Port:            8080,
			Ingresses:       []string{"app.example.com"},
			RoutingProvider: skiperatorv1alpha1.RoutingProviderStandard,
			ExtraContainers: []skiperatorv1alpha1.ContainerSpec{
				{Name: "auth-proxy", Image: "proxy:1.0", IngressPort: &proxyPort},
			},
		},
	}
	r := reconciliation.NewApplicationReconciliation(context.Background(), app, log.NewLogger(), false, nil, nil, config.SkiperatorConfig{})

	err := Generate(r)

	require.NoError(t, err)
	require.Len(t, r.GetResources(), 2)
	route := r.GetResources()[1].(*gatewayapiv1.HTTPRoute)
	assert.Equal(t, gatewayapiv1.ObjectName("app"), route.Spec.Rules[0].BackendRefs[0].Name)
	assert.Equal(t, gatewayapiv1.PortNumber(8080), *route.Spec.Rules[0].BackendRefs[0].Port)
}

func TestApplicationLegacyRoutingSkipsGatewayAPI(t *testing.T) {
	app := &skiperatorv1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{Name: "app", Namespace: "team-a"},
		Spec: skiperatorv1alpha1.ApplicationSpec{
			Image:           "image",
			Port:            8080,
			Ingresses:       []string{"app.example.com"},
			RoutingProvider: skiperatorv1alpha1.RoutingProviderLegacy,
		},
	}
	r := reconciliation.NewApplicationReconciliation(context.Background(), app, log.NewLogger(), false, nil, nil, config.SkiperatorConfig{})

	err := Generate(r)

	require.NoError(t, err)
	require.Empty(t, r.GetResources())
}

func skiperatorv1alpha1Bool(value bool) *bool {
	return &value
}
