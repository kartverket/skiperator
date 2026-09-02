package resourceprocessor

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestRequirePatch(t *testing.T) {
	depl := &v1.Deployment{}
	sa := &corev1.ServiceAccount{}

	deplShouldPatch := requirePatch(depl)
	saShouldPatch := requirePatch(sa)

	assert.True(t, deplShouldPatch)
	assert.False(t, saShouldPatch)
}

func TestPreserveReloaderStateKeepsAnnotationStrategyMarker(t *testing.T) {
	live := &corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				"reloader.stakater.com/last-reloaded-from": `[{"type":"SECRET","name":"azure-app","hash":"some-hash"}]`,
				"prometheus.io/scrape":                     "true",
			},
		},
	}
	desired := &corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{"prometheus.io/scrape": "true"},
		},
	}

	preserveReloaderState(desired, live)

	assert.Equal(t,
		`[{"type":"SECRET","name":"azure-app","hash":"some-hash"}]`,
		desired.Annotations["reloader.stakater.com/last-reloaded-from"],
	)
}
