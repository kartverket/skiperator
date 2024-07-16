package resourceschemas

import (
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	testing2 "testing"
)

func TestAddGVK(t *testing2.T) {
	//arrange
	scheme := runtime.NewScheme()
	corev1.AddToScheme(scheme)
	list := []client.ObjectList{&corev1.ServiceList{}}
	//act
	result := addGVKToList(list, scheme)
	//assert
	assert.NotEmpty(t, result)
	assert.NotEmpty(t, result[0].GroupVersionKind().Kind)
	assert.NotEmpty(t, result[0].GroupVersionKind().Version)
	assert.Equal(t, schema.GroupVersionKind{Group: "", Version: "v1", Kind: "ServiceList"}, result[0].GroupVersionKind())
}

func TestGetApplicationSchemas(t *testing2.T) {
	//arrange
	scheme := runtime.NewScheme()
	AddSchemas(scheme)
	//act
	result := GetApplicationSchemas(scheme)
	//assert
	assert.NotEmpty(t, result)
	assert.NotEmpty(t, result[0].GroupVersionKind().Kind)
	assert.NotEmpty(t, result[0].GroupVersionKind().Version)
	assert.Equal(t, schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "DeploymentList"}, result[0].GroupVersionKind())
}
