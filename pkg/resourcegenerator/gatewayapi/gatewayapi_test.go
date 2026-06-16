package gatewayapi

import (
	"context"
	"testing"
	"time"

	"github.com/kartverket/skiperator/api/common/istiotypes"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/internal/config"
	"github.com/kartverket/skiperator/pkg/gwapi"
	"github.com/kartverket/skiperator/pkg/log"
	"github.com/kartverket/skiperator/pkg/reconciliation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	gatewayapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

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

func TestRoutingStandardPathRewrite(t *testing.T) {
	routing := &skiperatorv1alpha1.Routing{
		ObjectMeta: metav1.ObjectMeta{Name: "api", Namespace: "team-a"},
		Spec: skiperatorv1alpha1.RoutingSpec{
			Hostname:        "api.example.com",
			RoutingProvider: skiperatorv1alpha1.RoutingProviderStandard,
			Routes: []skiperatorv1alpha1.Route{
				{TargetApp: "backend", PathPrefix: "/v1", RewriteUri: true, Port: 8080},
			},
		},
	}
	r := reconciliation.NewRoutingReconciliation(context.Background(), routing, log.NewLogger(), false, nil)

	err := Generate(r)

	require.NoError(t, err)
	require.Len(t, r.GetResources(), 3)

	listenerSet := r.GetResources()[0].(*gatewayapiv1.ListenerSet)
	assert.Equal(t, "team-a", listenerSet.Namespace)
	assert.Equal(t, gatewayapiv1.ObjectName(gwapi.ExternalGatewayName), listenerSet.Spec.ParentRef.Name)

	redirectRoute := r.GetResources()[1].(*gatewayapiv1.HTTPRoute)
	assert.Equal(t, "api-redirect", redirectRoute.Name)
	assert.Equal(t, httpSectionName, *redirectRoute.Spec.ParentRefs[0].SectionName)
	assert.Equal(t, gatewayapiv1.HTTPRouteFilterRequestRedirect, redirectRoute.Spec.Rules[0].Filters[0].Type)

	route := r.GetResources()[2].(*gatewayapiv1.HTTPRoute)
	assert.Equal(t, httpsSectionName, *route.Spec.ParentRefs[0].SectionName)
	assert.Equal(t, gatewayapiv1.HTTPRouteFilterURLRewrite, route.Spec.Rules[0].Filters[0].Type)
	assert.Equal(t, "/v1", *route.Spec.Rules[0].Matches[0].Path.Value)
	assert.Equal(t, gatewayapiv1.ObjectName("backend"), route.Spec.Rules[0].BackendRefs[0].Name)
	assert.Equal(t, gatewayapiv1.PortNumber(8080), *route.Spec.Rules[0].BackendRefs[0].Port)
}

func TestRetryPolicySkipsUnsupportedStringCodes(t *testing.T) {
	codes := []intstr.IntOrString{
		intstr.FromString("5xx"),
		intstr.FromString("retriable-4xx"),
		intstr.FromInt32(503),
	}
	attempts := int32(4)
	unsupportedOptions := make(map[string][]string)

	retry, err := retryPolicy(&istiotypes.Retries{
		Attempts:                 &attempts,
		RetryOnHttpResponseCodes: &codes,
	}, func(field string, value string) {
		unsupportedOptions[field] = append(unsupportedOptions[field], value)
	})

	require.NoError(t, err)
	require.NotNil(t, retry)
	require.Equal(t, 4, *retry.Attempts)
	require.Equal(t, []gatewayapiv1.HTTPRouteRetryStatusCode{503}, retry.Codes)
	require.Equal(t, []string{"5xx", "retriable-4xx"}, unsupportedOptions["retryOnHttpResponseCodes"])
}

func TestRetryPolicyDoesNotMapPerTryTimeoutToBackoff(t *testing.T) {
	timeout := metav1.Duration{Duration: 500 * time.Millisecond}
	unsupportedOptions := make(map[string][]string)

	retry, err := retryPolicy(&istiotypes.Retries{PerTryTimeout: &timeout}, func(field string, value string) {
		unsupportedOptions[field] = append(unsupportedOptions[field], value)
	})

	require.NoError(t, err)
	require.NotNil(t, retry)
	require.Nil(t, retry.Backoff)
	require.Equal(t, []string{"500ms"}, unsupportedOptions["perTryTimeout"])
}

func skiperatorv1alpha1Bool(value bool) *bool {
	return &value
}
