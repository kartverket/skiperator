package deployment

import (
	"github.com/kartverket/skiperator/pkg/testutil"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDeploymentMinimalApp(t *testing.T) {
	// Setup
	r := testutil.GetTestMinimalAppReconciliation()
	// Test
	err := Generate(r)

	// Assert
	assert.Nil(t, err)
	assert.Equal(t, 1, len(r.GetResources()))
}
