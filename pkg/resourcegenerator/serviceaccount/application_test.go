package serviceaccount

import (
	"github.com/kartverket/skiperator/pkg/testutil"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"testing"
)

func TestServiceAccountMinimalApp(t *testing.T) {
	// Setup
	r := testutil.GetTestMinimalAppReconciliation()
	// Test
	err := Generate(r)

	// Assert
	sa := r.GetResources()[0].(*corev1.ServiceAccount)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(r.GetResources()))
	assert.Equal(t, "minimal", sa.Name)
}
