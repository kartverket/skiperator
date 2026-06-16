package usage

import (
	"testing"

	commontypes "github.com/kartverket/skiperator/api/common"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestHasStalledRoutingMigration(t *testing.T) {
	obj := unstructured.Unstructured{Object: map[string]interface{}{
		"status": map[string]interface{}{
			"conditions": []interface{}{
				// Real stalled migrations set Ready and StandardRoutingReady.
				map[string]interface{}{
					"type":   commontypes.ReadyConditionType,
					"reason": migrationStalledReason,
				},
				map[string]interface{}{
					"type":   commontypes.StandardRoutingReadyConditionType,
					"reason": migrationStalledReason,
				},
			},
		},
	}}

	assert.True(t, hasStalledRoutingMigration(obj))
}

func TestHasStalledRoutingMigrationIgnoresReadyOnlyStalledCondition(t *testing.T) {
	obj := unstructured.Unstructured{Object: map[string]interface{}{
		"status": map[string]interface{}{
			"conditions": []interface{}{
				// Ready alone is not enough; metric must key off StandardRoutingReady.
				map[string]interface{}{
					"type":   commontypes.ReadyConditionType,
					"reason": migrationStalledReason,
				},
				map[string]interface{}{
					"type":   commontypes.StandardRoutingReadyConditionType,
					"reason": "StandardRoutingNotReady",
				},
			},
		},
	}}

	assert.False(t, hasStalledRoutingMigration(obj))
}
