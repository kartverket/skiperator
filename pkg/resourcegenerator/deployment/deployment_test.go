package deployment

import (
	"github.com/kartverket/skiperator/pkg/testutil"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	"testing"
)

func TestDeploymentMinimalAppShouldHaveLabels(t *testing.T) {
	// Setup
	r := testutil.GetTestMinimalAppReconciliation()
	// Test
	err := Generate(r)

	// Assert
	assert.Nil(t, err)
	assert.Equal(t, 1, len(r.GetResources()))
	depl := r.GetResources()[0].(*appsv1.Deployment)
	appLabel := map[string]string{"app": "minimal"}
	assert.Equal(t, appLabel["app"], depl.Spec.Selector.MatchLabels["app"])
	assert.Equal(t, appLabel["app"], depl.Spec.Template.Labels["app"])
}
