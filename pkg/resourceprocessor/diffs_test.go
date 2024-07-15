package resourceprocessor

import (
	"context"
	"github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/log"
	"github.com/kartverket/skiperator/pkg/reconciliation"
	"github.com/kartverket/skiperator/pkg/resourceschemas"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestGetDiffForApplication(t *testing.T) {
	scheme := runtime.NewScheme()
	resourceschemas.AddSchemas(scheme)
	mockClient := fake.NewClientBuilder().Build()
	resourceProcessor := NewResourceProcessor(mockClient, resourceschemas.GetApplicationSchemas(scheme), scheme)

	ctx := context.TODO()
	namespace := "test"
	labels := map[string]string{"app": "test-app"}

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
			Labels:    labels,
		},
	}

	newSA := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "new-sa",
			Namespace: namespace,
			Labels:    labels,
		},
	}
	// Create the live resource in the fake client
	err := mockClient.Create(ctx, liveSA)
	assert.Nil(t, err)
	r := reconciliation.NewApplicationReconciliation(context.TODO(), application, log.FromContext(context.Background()), nil, nil)
	var obj client.Object = newSA
	r.AddResource(&obj)
	shouldDelete, shouldCreate, shouldUpdate, err := resourceProcessor.getDiff(r)
	assert.Nil(t, err)
	assert.Len(t, shouldDelete, 1)
	assert.Len(t, shouldCreate, 1)
	assert.Len(t, shouldUpdate, 0)
}
