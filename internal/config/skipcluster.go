package config

import (
	"context"
	"errors"
	"github.com/kartverket/skiperator/pkg/util"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"net"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

type SKIPCluster struct {
	Name              string   `yaml:"name"`
	ControlPlaneCIDRs []string `yaml:"controlPlaneCIDRs"`
	WorkerNodeCIDRs   []string `yaml:"workerNodeCIDRs"`
}

type SKIPClusterList struct {
	Clusters []*SKIPCluster `yaml:"clusters"`
}

func (c *SKIPClusterList) CombinedCIDRS() []string {
	var combinedCIDRs []string
	for _, cluster := range c.Clusters {
		combinedCIDRs = append(combinedCIDRs, cluster.ControlPlaneCIDRs...)
		combinedCIDRs = append(combinedCIDRs, cluster.WorkerNodeCIDRs...)
	}
	return combinedCIDRs
}

func LoadConfigFromConfigMap(client client.Client) (*SKIPClusterList, error) {
	clusterConfigMapNamespaced := types.NamespacedName{Namespace: "skiperator-system", Name: "skip-cluster-node-cidr"}
	clusterConfigMap, err := util.GetConfigMap(client, context.Background(), clusterConfigMapNamespaced)
	if err != nil {
		return nil, err
	}
	clusterList, err := createSKIPClusterListFromConfigMap(&clusterConfigMap)
	if err != nil {
		return nil, err
	}
	return clusterList, nil
}

func createSKIPClusterListFromConfigMap(configMap *corev1.ConfigMap) (*SKIPClusterList, error) {
	var skipClusterList SKIPClusterList
	clusterData := configMap.Data

	// Safety check: make sure "config.yml" is present
	configYml, exists := clusterData["config.yml"]
	if !exists {
		return nil, errors.New("config.yml key not found in ConfigMap")
	}

	err := yaml.Unmarshal([]byte(configYml), &skipClusterList)
	if err != nil {
		return nil, err
	}
	if len(skipClusterList.Clusters) == 0 {
		return nil, errors.New("no SKIPClusterList in ConfigMap")
	}
	if len(skipClusterList.CombinedCIDRS()) == 0 {
		return nil, errors.New("no CIDR ranges in SKIPClusterList in ConfigMap")
	}

	// check for empty strings and validate that the strings are valid CIDRs
	for _, element := range skipClusterList.Clusters {
		err = checkCIDRsAreNotEmpty(element)
		if err != nil {
			return nil, err
		}
	}

	for _, element := range skipClusterList.CombinedCIDRS() {
		err = checkValidCIDR(element)
		if err != nil {
			return nil, err
		}
	}
	return &skipClusterList, nil
}

func checkCIDRsAreNotEmpty(cluster *SKIPCluster) error {
	if (len(cluster.WorkerNodeCIDRs) == 0) || (len(cluster.ControlPlaneCIDRs) == 0) {
		return errors.New("SKIPCluster has no CIDRs for worker nodes or control plane nodes")
	}
	for _, element := range cluster.WorkerNodeCIDRs {
		if element == "" {
			return errors.New("SKIPCluster has an empty worker node CIDR entry")
		}
	}
	for _, element := range cluster.ControlPlaneCIDRs {
		if element == "" {
			return errors.New("SKIPCluster has an empty control plane CIDR entry")
		}
	}
	return nil
}

func checkValidCIDR(cidr string) error {
	_, _, err := net.ParseCIDR(cidr)
	if err != nil {
		return err
	}
	return nil
}
