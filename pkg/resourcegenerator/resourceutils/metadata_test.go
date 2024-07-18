package resourceutils

import (
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
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
	// need to add gvk to find resource labels
	AddGVK(scheme.Scheme, sa)

	app := &skiperatorv1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testapp",
			Namespace: "testns",
			Labels:    map[string]string{"test": "test"},
		},
		Spec: skiperatorv1alpha1.ApplicationSpec{
			ResourceLabels: map[string]map[string]string{"ServiceAccount": {"someLabel": "someValue"}, "OtherResource": {"otherLabel": "otherValue"}},
		},
	}

	var obj client.Object = sa
	SetApplicationLabels(obj, app)
	assert.True(t, len(obj.GetLabels()) == 6)
	assert.True(t, obj.GetLabels()["someLabel"] == "someValue")
	assert.Empty(t, obj.GetLabels()["otherLabel"])
}
