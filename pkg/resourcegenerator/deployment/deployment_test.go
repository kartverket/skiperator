package deployment

import (
	"github.com/kartverket/skiperator/v2/api/v1alpha1"
	"github.com/kartverket/skiperator/v2/pkg/testutil"
	"github.com/kartverket/skiperator/v2/pkg/util"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
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

func TestHPAReplicasIsZero(t *testing.T) {
	// Setup
	r := testutil.GetTestMinimalAppReconciliation()
	r.GetSKIPObject().(*v1alpha1.Application).Spec.Replicas = &apiextensionsv1.JSON{Raw: []byte(`{"min": 0, "max": 0, "targetCpuUtilization": 200, "targetMemoryUtilization": 200}`)}
	// Test
	err := Generate(r)

	// Assert
	assert.Nil(t, err)
	depl := r.GetResources()[0].(*appsv1.Deployment)
	assert.Equal(t, depl.Spec.Replicas, util.PointTo(int32(0)))
}
