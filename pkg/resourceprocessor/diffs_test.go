package resourceprocessor

import (
	"context"
	"testing"

	"github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/internal/config"
	"github.com/kartverket/skiperator/pkg/log"
	"github.com/kartverket/skiperator/pkg/reconciliation"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/resourceutils"
	"github.com/kartverket/skiperator/pkg/resourceschemas"
	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestGetDiffForApplicationShouldCreateDelete(t *testing.T) {
	scheme := runtime.NewScheme()
	resourceschemas.AddSchemas(scheme)
	mockClient := fake.NewClientBuilder().Build()
	resourceProcessor := NewResourceProcessor(mockClient, resourceschemas.GetApplicationSchemas(scheme), scheme)

	ctx := context.TODO()
	namespace := "test"

	application := &v1alpha1.Application{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-app",
			Namespace: namespace,
		},
	}
	liveSA := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "live-sa",
			Namespace: namespace,
			Labels:    application.GetDefaultLabels(),
		},
	}
	newSA := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "new-sa",
			Namespace: namespace,
			Labels:    application.GetDefaultLabels(),
		},
	}

	ignoreLabels := application.GetDefaultLabels()
	ignoreLabels["skiperator.kartverket.no/ignore"] = "true"
	liveDeploymentDontDelete := &v1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-app",
			Namespace: namespace,
			Labels:    ignoreLabels,
		},
	}
	liveDeploymentIgnorePatchOrCreate := &v1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-app2",
			Namespace: namespace,
			Labels:    ignoreLabels,
		},
	}

	// Create the live resource in the fake client
	err := mockClient.Create(ctx, liveDeploymentDontDelete)
	assert.Nil(t, err)
	err = mockClient.Create(ctx, liveDeploymentIgnorePatchOrCreate)
	assert.Nil(t, err)
	err = mockClient.Create(ctx, liveSA)
	assert.Nil(t, err)
	r := reconciliation.NewApplicationReconciliation(context.TODO(), application, log.NewLogger(), false, nil, nil, config.SkiperatorConfig{
		TopologyKeys:                nil,
		LeaderElection:              false,
		LeaderElectionNamespace:     "",
		ConcurrentReconciles:        0,
		IsDeployment:                false,
		LogLevel:                    "",
		EnableProfiling:             false,
		RegistryCredentials:         nil,
		ClusterCIDRExclusionEnabled: false,
		ClusterCIDRMap:              config.SKIPClusterList{},
		EnableLocallyBuiltImages:    false,
		GCPIdentityProvider:         "",
		GCPWorkloadIdentityPool:     "",
	})
	resourceutils.AddGVK(scheme, newSA)
	resourceutils.AddGVK(scheme, liveDeploymentIgnorePatchOrCreate)
	//build reconcile objects array
	r.AddResource(newSA)
	r.AddResource(liveDeploymentIgnorePatchOrCreate)
	diffs, err := resourceProcessor.getDiff(r)
	assert.Nil(t, err)
	assert.Len(t, diffs.shouldDelete, 1)
	assert.Len(t, diffs.shouldCreate, 1)
	assert.Len(t, diffs.shouldUpdate, 0)
	assert.Len(t, diffs.shouldPatch, 0)
}

func TestCompareObjectShouldEqual(t *testing.T) {
	sa1 := &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ServiceAccount",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-sa",
			Namespace: "test",
		},
	}
	sa2 := &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ServiceAccount",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-sa",
			Namespace: "test",
		},
	}

	isEqual := compareObject(sa1, sa2)
	assert.True(t, isEqual)
}

func TestCompareObjectShouldNotEqualNamespace(t *testing.T) {
	sa1 := &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ServiceAccount",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-sa",
			Namespace: "test",
		},
	}
	sa2 := &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ServiceAccount",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-sa",
			Namespace: "test2",
		},
	}

	isEqual := compareObject(sa1, sa2)
	assert.False(t, isEqual)
}

func TestCompareObjectShouldNotEqualName(t *testing.T) {
	sa1 := &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ServiceAccount",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-sa",
			Namespace: "test",
		},
	}
	sa2 := &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ServiceAccount",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-sa2",
			Namespace: "test",
		},
	}

	isEqual := compareObject(sa1, sa2)
	assert.False(t, isEqual)
}

func TestCompareObjectShouldNotEqualType(t *testing.T) {
	sa := &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ServiceAccount",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-sa",
			Namespace: "test",
		},
	}
	configMap := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-cm",
			Namespace: "test",
		},
	}

	isEqual := compareObject(sa, configMap)
	assert.False(t, isEqual)
}
