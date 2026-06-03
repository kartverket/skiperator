package v1alpha1

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestApplicationFillDefaultsStatusIsIdempotent(t *testing.T) {
	app := &Application{}

	app.FillDefaultsStatus()
	firstStatus := app.Status.DeepCopy()

	time.Sleep(time.Nanosecond)
	app.FillDefaultsStatus()

	assert.Equal(t, firstStatus.Summary, app.Status.Summary)
	assert.NotNil(t, app.Status.SubResources)
	assert.NotNil(t, app.Status.Conditions)
}
