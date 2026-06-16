package certificate

import (
	"context"
	"testing"

	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/internal/config"
	"github.com/kartverket/skiperator/pkg/log"
	"github.com/kartverket/skiperator/pkg/reconciliation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestApplicationStandardRoutingGeneratesOnlyLocalCertWhenLegacyDisabled(t *testing.T) {
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
	require.Len(t, r.GetResources(), 1)
	assert.Equal(t, "team-a", r.GetResources()[0].GetNamespace())
}

func TestApplicationStandardRoutingKeepsLegacyCertWhenLegacyEnabled(t *testing.T) {
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

	err := Generate(r)

	require.NoError(t, err)
	require.Len(t, r.GetResources(), 2)
	assert.Equal(t, IstioGatewayNamespace, r.GetResources()[0].GetNamespace())
	assert.Equal(t, "team-a", r.GetResources()[1].GetNamespace())
}
