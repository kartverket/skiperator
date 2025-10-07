package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidSKIPCluster(t *testing.T) {

	skipClusterList := SKIPClusterList{
		Clusters: []*SKIPCluster{
			{
				Name:              "",
				ControlPlaneCIDRs: []string{"1.1.1.1/32", "2.2.2.2/32"},
				WorkerNodeCIDRs:   []string{"1.1.1.1/32", "2.2.2.2/32"},
			},
		},
	}
	err := ValidateSKIPClusterList(&skipClusterList)
	assert.NoError(t, err)
}

func TestInvalidCIDRThrowsError(t *testing.T) {
	skipClusterList := SKIPClusterList{
		Clusters: []*SKIPCluster{
			{
				Name:              "",
				ControlPlaneCIDRs: []string{"11221"},
				WorkerNodeCIDRs:   []string{"111"},
			},
		},
	}
	err := ValidateSKIPClusterList(&skipClusterList)
	assert.EqualErrorf(t, err, "invalid CIDR address: 11221", "11221")
}

func TestNoCIDRRangesInConfigMap(t *testing.T) {
	skipClusterList := SKIPClusterList{
		Clusters: []*SKIPCluster{
			{
				Name:              "",
				ControlPlaneCIDRs: []string{},
				WorkerNodeCIDRs:   []string{},
			},
		},
	}

	err := ValidateSKIPClusterList(&skipClusterList)
	assert.EqualErrorf(t, err, "no CIDR ranges in SKIPClusterList in ConfigMap", "")
}

func TestControlPlaneCIDRsAreMissingOK(t *testing.T) {
	skipClusterList := SKIPClusterList{
		Clusters: []*SKIPCluster{
			{
				Name:              "",
				ControlPlaneCIDRs: []string{},
				WorkerNodeCIDRs:   []string{"1.2.3.4/27", "1.2.3.4/28"},
			},
		},
	}
	err := ValidateSKIPClusterList(&skipClusterList)
	assert.NoError(t, err)
}

func TestWorkerNodeCIDRsAreMissingOK(t *testing.T) {
	skipClusterList := SKIPClusterList{
		Clusters: []*SKIPCluster{
			{
				Name:              "",
				ControlPlaneCIDRs: []string{"1.2.3.4/25"},
			},
		},
	}

	err := ValidateSKIPClusterList(&skipClusterList)
	assert.NoError(t, err)
}

func TestEmptyControlPlaneCIDREntryThrowsError(t *testing.T) {
	skipClusterList := SKIPClusterList{
		Clusters: []*SKIPCluster{
			{
				Name:              "",
				ControlPlaneCIDRs: []string{"1.2.3.4/25", ""},
				WorkerNodeCIDRs:   []string{"1.2.3.4/27", "1.2.3.4/28"},
			},
		},
	}

	err := ValidateSKIPClusterList(&skipClusterList)
	assert.EqualErrorf(t, err, "SKIPCluster has an empty control plane CIDR entry", "")
}

func TestEmptyWorkerNodeCIDREntryThrowsError(t *testing.T) {
	skipClusterList := SKIPClusterList{
		Clusters: []*SKIPCluster{
			{
				Name:              "",
				ControlPlaneCIDRs: []string{"1.2.3.4/25", "1.2.3.4/26"},
				WorkerNodeCIDRs:   []string{"1.2.3.4/27", ""},
			},
		},
	}

	err := ValidateSKIPClusterList(&skipClusterList)
	assert.EqualErrorf(t, err, "SKIPCluster has an empty worker node CIDR entry", "")
}

func TestEmptyConfigMapThrowsError(t *testing.T) {
	skipClusterList := SKIPClusterList{}
	err := ValidateSKIPClusterList(&skipClusterList)
	assert.EqualErrorf(t, err, "no SKIPClusterList in ConfigMap", "")
}
