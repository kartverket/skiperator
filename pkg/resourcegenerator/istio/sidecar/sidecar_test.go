package sidecar

import (
	"context"
	"fmt"
	"testing"

	"github.com/kartverket/skiperator/api/common/podtypes"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	skiperatorv1beta1 "github.com/kartverket/skiperator/api/v1beta1"
	"github.com/kartverket/skiperator/internal/config"
	"github.com/kartverket/skiperator/pkg/log"
	"github.com/kartverket/skiperator/pkg/reconciliation"
	"github.com/kartverket/skiperator/pkg/testutil"
	"github.com/kartverket/skiperator/pkg/util"
	"github.com/stretchr/testify/assert"
	networkingv1api "istio.io/api/networking/v1"
	networkingv1 "istio.io/client-go/pkg/apis/networking/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNamespaceGeneratesRegistryOnlySidecar(t *testing.T) {
	namespace := skiperatorv1alpha1.SKIPNamespace{
		Namespace: &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{Name: "test"},
		},
	}
	r := reconciliation.NewNamespaceReconciliation(context.TODO(), namespace, log.NewLogger(), false, nil)

	err := Generate(r)

	assert.NoError(t, err)
	assert.Len(t, r.GetResources(), 1)
	sidecar := r.GetResources()[0].(*networkingv1.Sidecar)
	assert.Equal(t, "test", sidecar.Namespace)
	assert.Equal(t, "sidecar", sidecar.Name)
	assert.Nil(t, sidecar.Spec.WorkloadSelector)
	assert.Nil(t, sidecar.Spec.Egress)
	assert.Equal(t, networkingv1api.OutboundTrafficPolicy_REGISTRY_ONLY, sidecar.Spec.OutboundTrafficPolicy.Mode)
}

func TestApplicationWithoutOutboundImportsNoHosts(t *testing.T) {
	r := testutil.GetTestMinimalAppReconciliation()

	err := Generate(r)

	assert.NoError(t, err)
	assert.Len(t, r.GetResources(), 1)
	sidecar := r.GetResources()[0].(*networkingv1.Sidecar)
	assert.Equal(t, "minimal", sidecar.Name)
	assert.Equal(t, map[string]string{"app": "minimal"}, sidecar.Spec.WorkloadSelector.Labels)
	assert.Equal(t, []string{"~/*"}, sidecar.Spec.Egress[0].Hosts)
}

func TestSKIPJobWithoutOutboundUsesSkipJobWorkloadName(t *testing.T) {
	skipJob := &skiperatorv1beta1.SKIPJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "minimal",
			Namespace: "test",
		},
		Spec: skiperatorv1beta1.SKIPJobSpec{
			Image: "image",
		},
	}
	r := reconciliation.NewJobReconciliation(context.TODO(), skipJob, log.NewLogger(), false, nil, config.SkiperatorConfig{})

	err := Generate(r)

	assert.NoError(t, err)
	assert.Len(t, r.GetResources(), 1)
	sidecar := r.GetResources()[0].(*networkingv1.Sidecar)
	assert.Equal(t, "minimal-skipjob", sidecar.Name)
	assert.Equal(t, map[string]string{"app": "minimal-skipjob"}, sidecar.Spec.WorkloadSelector.Labels)
	assert.Equal(t, []string{"~/*"}, sidecar.Spec.Egress[0].Hosts)
}

func TestApplicationWithExternalOutboundImportsOnlyDeclaredHost(t *testing.T) {
	r := testutil.GetTestMinimalAppReconciliation()
	application := r.GetSKIPObject().(*skiperatorv1alpha1.Application)
	application.Spec.AccessPolicy = &podtypes.AccessPolicy{
		Outbound: &podtypes.OutboundPolicy{
			External: []podtypes.ExternalRule{
				{Host: "www.vg.no"},
			},
		},
	}

	err := Generate(r)

	assert.NoError(t, err)
	assert.Len(t, r.GetResources(), 1)
	sidecar := r.GetResources()[0].(*networkingv1.Sidecar)
	assert.Equal(t, []string{"./www.vg.no"}, sidecar.Spec.Egress[0].Hosts)
}

func TestApplicationWithCloudSQLProxyImportsGeneratedServiceEntryHost(t *testing.T) {
	r := testutil.GetTestMinimalAppReconciliation()
	application := r.GetSKIPObject().(*skiperatorv1alpha1.Application)
	application.Spec.GCP = &podtypes.GCP{
		CloudSQLProxy: &podtypes.CloudSQLProxySettings{},
	}

	err := Generate(r)

	assert.NoError(t, err)
	assert.Len(t, r.GetResources(), 1)
	sidecar := r.GetResources()[0].(*networkingv1.Sidecar)
	assert.Equal(t, []string{
		fmt.Sprintf("./minimal-%x.cloudsql", util.GenerateHashFromName("image")),
	}, sidecar.Spec.Egress[0].Hosts)
}

func TestApplicationWithInternalOutboundImportsServiceFQDN(t *testing.T) {
	r := testutil.GetTestMinimalAppReconciliation()
	r.SetResolvedOutboundRules([]reconciliation.ResolvedOutboundRule{
		{Application: "dependency", Namespaces: []string{"another"}},
	})
	application := r.GetSKIPObject().(*skiperatorv1alpha1.Application)
	application.Spec.AccessPolicy = &podtypes.AccessPolicy{
		Outbound: &podtypes.OutboundPolicy{
			Rules: []podtypes.InternalRule{
				{Application: "dependency", Namespace: "another"},
			},
		},
	}

	err := Generate(r)

	assert.NoError(t, err)
	assert.Len(t, r.GetResources(), 1)
	sidecar := r.GetResources()[0].(*networkingv1.Sidecar)
	assert.Equal(t, []string{"another/dependency.another.svc.cluster.local"}, sidecar.Spec.Egress[0].Hosts)
}

func TestApplicationWithNamespacesByLabelOutboundImportsResolvedNamespaces(t *testing.T) {
	r := testutil.GetTestMinimalAppReconciliation()
	r.SetResolvedOutboundRules([]reconciliation.ResolvedOutboundRule{
		{Application: "dependency", Namespaces: []string{"ateam-main", "ateam-feat"}},
	})
	application := r.GetSKIPObject().(*skiperatorv1alpha1.Application)
	application.Spec.AccessPolicy = &podtypes.AccessPolicy{
		Outbound: &podtypes.OutboundPolicy{
			Rules: []podtypes.InternalRule{
				{Application: "dependency", NamespacesByLabel: map[string]string{"team": "a"}},
			},
		},
	}

	err := Generate(r)

	assert.NoError(t, err)
	assert.Len(t, r.GetResources(), 1)
	sidecar := r.GetResources()[0].(*networkingv1.Sidecar)
	assert.Equal(t, []string{
		"ateam-feat/dependency.ateam-feat.svc.cluster.local",
		"ateam-main/dependency.ateam-main.svc.cluster.local",
	}, sidecar.Spec.Egress[0].Hosts)
}
