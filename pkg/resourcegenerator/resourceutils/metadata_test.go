package resourceutils

import (
	"testing"

	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
)

func TestSetResourceLabels(t *testing.T) {
	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "testns",
		},
	}
	// need to add gvk to find resource labels
	_ = AddGVK(scheme.Scheme, sa)

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

	expectedLabels := map[string]string{
		"app.kubernetes.io/name":                  "testapp",
		"app.kubernetes.io/managed-by":            "skiperator",
		"skiperator.kartverket.no/controller":     "application",
		"application.skiperator.no/app":           "testapp",
		"application.skiperator.no/app-name":      "testapp",
		"application.skiperator.no/app-namespace": "testns",
		"someLabel": "someValue",
	}

	SetApplicationLabels(sa, app)
	assert.Equal(t, expectedLabels, sa.GetLabels())
	assert.Empty(t, sa.GetLabels()["otherLabel"])
}
