package common

import (
	"github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/testutil"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestShouldReconcile(t *testing.T) {
	r := testutil.GetTestMinimalAppReconciliation()
	app := r.GetReconciliationObject().(*v1alpha1.Application)
	assert.True(t, ShouldReconcile(app))
	app.Labels["skiperator.kartverket.no/ignore"] = "true"
	assert.False(t, ShouldReconcile(app))
}
