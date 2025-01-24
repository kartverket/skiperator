package common

import (
	"github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/testutil"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
	"time"
)

func TestShouldReconcile(t *testing.T) {
	r := testutil.GetTestMinimalAppReconciliation()
	app := r.GetSKIPObject().(*v1alpha1.Application)
	assert.True(t, ShouldReconcile(app))
	app.Labels["skiperator.kartverket.no/ignore"] = "true"
	assert.False(t, ShouldReconcile(app))
}

func TestStatusDiffWithTimestamp(t *testing.T) {
	status := &v1alpha1.SkiperatorStatus{
		Summary: v1alpha1.Status{
			Status:    v1alpha1.SYNCED,
			Message:   "All subresources synced",
			TimeStamp: time.Now().String(),
		},
		Conditions: []v1.Condition{
			{
				ObservedGeneration: 1,
				LastTransitionTime: v1.Now(),
			},
		},
		SubResources: map[string]v1alpha1.Status{
			"test": {
				Status:    v1alpha1.SYNCED,
				Message:   "All subresources synced",
				TimeStamp: time.Now().String(),
			},
		},
	}
	tmpStatus := status.DeepCopy()
	status.Summary.TimeStamp = time.Now().String()
	status.Conditions[0].LastTransitionTime = v1.Now()
	status.SubResources["test"] = v1alpha1.Status{
		Status:    v1alpha1.SYNCED,
		Message:   "All subresources synced",
		TimeStamp: time.Now().String(),
	}

	//assert that timestamps are in fact different
	assert.NotEqual(t, tmpStatus.Summary.TimeStamp, status.Summary.TimeStamp)
	assert.NotEqual(t, tmpStatus.Conditions[0].LastTransitionTime, status.Conditions[0].LastTransitionTime)
	assert.NotEqual(t, tmpStatus.SubResources["test"].TimeStamp, status.SubResources["test"].TimeStamp)

	//assert zero diff
	diff, err := GetObjectDiff(tmpStatus, status)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(diff))
}
