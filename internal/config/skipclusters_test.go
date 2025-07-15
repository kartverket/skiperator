package config

import (
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"testing"
)

func TestValidConfigMap(t *testing.T) {
	mapData := make(map[string]string)
	mapData["config.yml"] = "clusters: \n  - name: \"test-cluster-1\"\n    controlPlaneCIDRs:\n      - 1.2.3.4/25\n      - 1.2.3.4/26\n    workerNodeCIDRs:\n      - 1.2.3.4/27\n      - 1.2.3.4/28\n\n    \n"
	configMap := v1.ConfigMap{
		Data: mapData,
	}
	clusterList, _ := createSKIPClusterListFromConfigMap(&configMap)

	assert.NotNil(t, clusterList)
	assert.Equal(t, 1, len(clusterList.Clusters))
	assert.Equal(t, "test-cluster-1", clusterList.Clusters[0].Name)
	assert.Equal(t, 2, len(clusterList.Clusters[0].ControlPlaneCIDRs))
	assert.Equal(t, 2, len(clusterList.Clusters[0].WorkerNodeCIDRs))
	assert.Equal(t, "1.2.3.4/25", clusterList.Clusters[0].ControlPlaneCIDRs[0])
	assert.Equal(t, "1.2.3.4/27", clusterList.Clusters[0].WorkerNodeCIDRs[0])
}

func TestInvalidConfigMapThrowsError(t *testing.T) {
	mapData := make(map[string]string)
	mapData["config.yml"] = "clusters: \n  - name: \"test-cluster-1\"\n    controlPlaneCIDRs:\n      - 11221\n    workerNodeCIDRs:\n      - 111\n\n    \n"
	configMap := v1.ConfigMap{
		Data: mapData,
	}
	_, err := createSKIPClusterListFromConfigMap(&configMap)
	assert.EqualErrorf(t, err, "invalid CIDR address: 11221", "11221")

}

func TestNoCIDRRangesInConfigMap(t *testing.T) {
	mapData := make(map[string]string)
	mapData["config.yml"] = "clusters: \n  - name: \"test-cluster-1\"\n    controlPlaneCIDRs:\n    workerNodeCIDRs:"
	configMap := v1.ConfigMap{
		Data: mapData,
	}
	_, err := createSKIPClusterListFromConfigMap(&configMap)
	assert.EqualErrorf(t, err, "no CIDR ranges in SKIPClusterList in ConfigMap", "")
}

func TestControlPlaneCIDRsAreMissing(t *testing.T) {
	mapData := make(map[string]string)
	mapData["config.yml"] = "clusters: \n  - name: \"test-cluster-1\"\n    controlPlaneCIDRs:\n    workerNodeCIDRs:\n      - 1.2.3.4/27\n      - 1.2.3.4/28\n\n    \n"
	configMap := v1.ConfigMap{
		Data: mapData,
	}
	_, err := createSKIPClusterListFromConfigMap(&configMap)
	assert.EqualErrorf(t, err, "SKIPCluster has no CIDRs for worker nodes or control plane nodes", "")
}

func TestWorkerNodeCIDRsAreMissing(t *testing.T) {
	mapData := make(map[string]string)
	mapData["config.yml"] = "clusters: \n  - name: \"test-cluster-1\"\n    controlPlaneCIDRs:\n      - 1.2.3.4/25\n      - \"\"\n    workerNodeCIDRs:\n"
	configMap := v1.ConfigMap{
		Data: mapData,
	}
	_, err := createSKIPClusterListFromConfigMap(&configMap)
	assert.EqualErrorf(t, err, "SKIPCluster has no CIDRs for worker nodes or control plane nodes", "")
}

func TestNoConfigKeyInConfigMapThrowsError(t *testing.T) {
	mapData := make(map[string]string)

	configMap := v1.ConfigMap{
		Data: mapData,
	}
	_, err := createSKIPClusterListFromConfigMap(&configMap)
	assert.EqualErrorf(t, err, "config.yml key not found in ConfigMap", "")
}

func TestEmptyConfigMapThrowsError(t *testing.T) {
	mapData := make(map[string]string)
	mapData["config.yml"] = ""
	configMap := v1.ConfigMap{
		Data: mapData,
	}
	_, err := createSKIPClusterListFromConfigMap(&configMap)
	assert.EqualErrorf(t, err, "no SKIPClusterList in ConfigMap", "")
}

func TestEmptyControlPlaneCIDREntryThrowsError(t *testing.T) {
	mapData := make(map[string]string)
	mapData["config.yml"] = "clusters: \n  - name: \"test-cluster-1\"\n    controlPlaneCIDRs:\n      - 1.2.3.4/25\n      - \"\"\n    workerNodeCIDRs:\n      - 1.2.3.4/27\n      - 1.2.3.4/28\n\n    \n"
	configMap := v1.ConfigMap{
		Data: mapData,
	}
	_, err := createSKIPClusterListFromConfigMap(&configMap)
	assert.EqualErrorf(t, err, "SKIPCluster has an empty control plane CIDR entry", "")
}

func TestEmptyWorkerNodeCIDREntryThrowsError(t *testing.T) {
	mapData := make(map[string]string)
	mapData["config.yml"] = "clusters: \n  - name: \"test-cluster-1\"\n    controlPlaneCIDRs:\n      - 1.2.3.4/25\n    workerNodeCIDRs:\n      - 1.2.3.4/27\n      - \"\"\n\n    \n"
	configMap := v1.ConfigMap{
		Data: mapData,
	}
	_, err := createSKIPClusterListFromConfigMap(&configMap)
	assert.EqualErrorf(t, err, "SKIPCluster has an empty worker node CIDR entry", "")
}
