package resourceutils

import (
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"testing"
)

func TestSetResourceLabels(t *testing.T) {
	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "testns",
		},
	}

	app := &skiperatorv1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testapp",
			Namespace: "testns",
		},
		Spec: skiperatorv1alpha1.ApplicationSpec{
			ResourceLabels: map[string]map[string]string{"ServiceAccount": {"someLabel": "someValue"}, "OtherResource": {"otherLabel": "otherValue"}},
		},
	}

	var obj client.Object = sa
	setResourceLabels(obj, app)
	assert.True(t, len(obj.GetLabels()) == 1)
	assert.True(t, obj.GetLabels()["someLabel"] == "someValue")
	assert.Nil(t, obj.GetLabels()["otherLabel"])
}
