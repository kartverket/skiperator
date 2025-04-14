package common

import (
	"testing"
	"time"

	"github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/api/v1alpha1/podtypes"
	"github.com/kartverket/skiperator/pkg/testutil"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func TestShouldNormalizeHosts(t *testing.T) {
	// Empty cases
	t.Run("empty_cases", func(t *testing.T) {
		assert.True(t, IsExternalRulesValid(nil))
		assert.True(t, IsExternalRulesValid(&podtypes.AccessPolicy{}))
	})

	// Valid cases
	t.Run("single_domain", func(t *testing.T) {
		singleDomain := externalPolicyTo(
			podtypes.ExternalRule{
				Host: "foo.com",
				Ports: []podtypes.ExternalPort{
					{Port: 80, Name: "http"},
					{Port: 443, Name: "https"},
				},
			},
		)
		assert.True(t, IsExternalRulesValid(singleDomain))
	})

	// Invalid cases
	t.Run("duplicate_domain", func(t *testing.T) {
		duplicateDomain := externalPolicyTo(
			podtypes.ExternalRule{
				Host: "foo.com",
				Ports: []podtypes.ExternalPort{
					{Port: 80, Name: "http"},
				},
			},
			podtypes.ExternalRule{
				Host: "foo.com",
				Ports: []podtypes.ExternalPort{
					{Port: 443, Name: "https"},
				},
			},
		)
		assert.False(t, IsExternalRulesValid(duplicateDomain))
	})

	t.Run("duplicate_domain_with_casing", func(t *testing.T) {
		duplicateDomainWithCasing := externalPolicyTo(
			podtypes.ExternalRule{
				Host: "FoO.CoM",
				Ports: []podtypes.ExternalPort{
					{Port: 80, Name: "http"},
				},
			},
			podtypes.ExternalRule{
				Host: "foo.com",
				Ports: []podtypes.ExternalPort{
					{Port: 443, Name: "https"},
				},
			},
		)
		assert.False(t, IsExternalRulesValid(duplicateDomainWithCasing))
	})

	t.Run("bad_hostname", func(t *testing.T) {
		badHostname := externalPolicyTo(
			podtypes.ExternalRule{
				Host: "foo.999",
				Ports: []podtypes.ExternalPort{
					{Port: 80, Name: "http"},
				},
			},
		)
		assert.False(t, IsExternalRulesValid(badHostname))
	})

	t.Run("protocol_in_host", func(t *testing.T) {
		protocolInHost := externalPolicyTo(
			podtypes.ExternalRule{
				Host: "http://foo.com",
				Ports: []podtypes.ExternalPort{
					{Port: 80, Name: "http"},
				},
			},
		)
		assert.False(t, IsExternalRulesValid(protocolInHost))
	})
}

func externalPolicyTo(rules ...podtypes.ExternalRule) *podtypes.AccessPolicy {
	return &podtypes.AccessPolicy{
		Outbound: &podtypes.OutboundPolicy{
			External: rules,
		},
	}
}
