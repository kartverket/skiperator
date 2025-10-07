package config

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var emptyConfig = SkiperatorConfig{}

func TestDefaultConfig(t *testing.T) {
	cm := mockConfigMap(emptyConfig, nil)
	err := parseConfig(cm)

	assert.NoError(t, err)
	assert.Equal(t, 1, len(GetActiveConfig().TopologyKeys))
	assert.Contains(t, GetActiveConfig().TopologyKeys, "kubernetes.io/hostname")
}

func TestCustomTopologyKeys(t *testing.T) {
	cm := mockConfigMap(SkiperatorConfig{
		TopologyKeys: []string{
			"kubernetes.io/hostname",
			"topology.kubernetes.io/zone",
			"skip.kartverket.no/unit-test",
		},
	}, nil)
	err := parseConfig(cm)

	assert.NoError(t, err)
	assert.Equal(t, 3, len(GetActiveConfig().TopologyKeys))
	assert.Contains(t, GetActiveConfig().TopologyKeys, "kubernetes.io/hostname")
	assert.Contains(t, GetActiveConfig().TopologyKeys, "topology.kubernetes.io/zone")
	assert.Contains(t, GetActiveConfig().TopologyKeys, "skip.kartverket.no/unit-test")
}

func TestUnknownKey(t *testing.T) {
	cm := mockConfigMap(emptyConfig, map[string]any{
		"someUnknownKey": "someValue",
	})
	err := parseConfig(cm)

	assert.Error(t, err)
}

func TestBadlyFormattedConfig(t *testing.T) {
	cases := []struct {
		name              string
		topologyKeysValue any
	}{
		{"string", "it-should-be-an-array"},
		{"number", 42},
		{"boolean", true},
		{"array-of-boolean", []bool{true, false}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cm := mockConfigMap(emptyConfig, map[string]any{
				"topologyKeys": tc.topologyKeysValue,
			})
			err := parseConfig(cm)

			assert.Error(t, err, "Expected error for topologyKeys with value: %v", tc.topologyKeysValue)
		})
	}
}

func TestCompleteConfigMap(t *testing.T) {
	cm := mockConfigMap(SkiperatorConfig{
		TopologyKeys: []string{
			"kubernetes.io/hostname",
			"topology.kubernetes.io/zone",
			"skip.kartverket.no/unit-test",
		},
		LeaderElection:          true,
		LeaderElectionNamespace: "skiperator-system",
		ConcurrentReconciles:    1,
		IsDeployment:            true,
		LogLevel:                "debug",
		RegistryCredentials: []RegistryCredentialPair{
			RegistryCredentialPair{
				Registry: "foo",
				Token:    "bar",
			},
		},
		ClusterCIDRExclusionEnabled: false,
		ClusterCIDRMap: SKIPClusterList{
			Clusters: []*SKIPCluster{
				{
					Name:              "",
					ControlPlaneCIDRs: []string{"1.1.1.1/32", "2.2.2.2/32"},
					WorkerNodeCIDRs:   []string{"1.1.1.1/32", "2.2.2.2/32"},
				},
			},
		},
		EnableLocallyBuiltImages: true,
		GCPIdentityProvider:      "foobar",
		GCPWorkloadIdentityPool:  "barfoo",
	}, nil)
	err := parseConfig(cm)

	assert.NoError(t, err)
	assert.Equal(t, 3, len(GetActiveConfig().TopologyKeys))
	assert.Contains(t, GetActiveConfig().TopologyKeys, "kubernetes.io/hostname")
	assert.Contains(t, GetActiveConfig().TopologyKeys, "topology.kubernetes.io/zone")
	assert.Contains(t, GetActiveConfig().TopologyKeys, "skip.kartverket.no/unit-test")
	assert.Equal(t, GetActiveConfig().IsDeployment, true)
	assert.Equal(t, 1, len(GetActiveConfig().RegistryCredentials))
	assert.Equal(t, "foo", GetActiveConfig().RegistryCredentials[0].Registry)
	assert.Equal(t, "bar", GetActiveConfig().RegistryCredentials[0].Token)
	assert.Equal(t, 1, len(GetActiveConfig().ClusterCIDRMap.Clusters))
	assert.Equal(t, 2, len(GetActiveConfig().ClusterCIDRMap.Clusters[0].ControlPlaneCIDRs))
	assert.Equal(t, 2, len(GetActiveConfig().ClusterCIDRMap.Clusters[0].WorkerNodeCIDRs))
	assert.Equal(t, "1.1.1.1/32", GetActiveConfig().ClusterCIDRMap.Clusters[0].WorkerNodeCIDRs[0])
	assert.Equal(t, true, GetActiveConfig().EnableLocallyBuiltImages)
	assert.Equal(t, "foobar", GetActiveConfig().GCPIdentityProvider)
	assert.Equal(t, "barfoo", GetActiveConfig().GCPWorkloadIdentityPool)

}

func mockConfigMap(c SkiperatorConfig, extra map[string]any) *v1.ConfigMap {
	jsonPayload, _ := json.Marshal(c)

	if extra != nil {
		var dump map[string]any
		_ = json.Unmarshal(jsonPayload, &dump)
		for k, v := range extra {
			dump[k] = v
		}
		jsonPayload, _ = json.Marshal(dump)
	}

	return &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cmName,
			Namespace: cmNamespace,
		},
		Data: map[string]string{
			cmKey: string(jsonPayload),
		},
	}
}
