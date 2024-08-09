package resourceprocessor

import (
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"testing"
)

func TestRequirePatch(t *testing.T) {
	depl := &v1.Deployment{}
	sa := &corev1.ServiceAccount{}

	deplShouldPatch := requirePatch(depl)
	saShouldPatch := requirePatch(sa)

	assert.True(t, deplShouldPatch)
	assert.False(t, saShouldPatch)
}
