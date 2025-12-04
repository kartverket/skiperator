package config

import (
	"errors"
	"net"
	"strings"
)

type SKIPCluster struct {
	Name              string   `json:"name"`
	ControlPlaneCIDRs []string `json:"controlPlaneCIDRs,omitempty"`
	WorkerNodeCIDRs   []string `json:"workerNodeCIDRs,omitempty"`
}

type SKIPClusterList struct {
	Clusters []*SKIPCluster `json:"clusters,omitempty"`
}

func (c *SKIPClusterList) CombinedCIDRS() []string {
	var combinedCIDRs []string
	for _, cluster := range c.Clusters {
		combinedCIDRs = append(combinedCIDRs, cluster.ControlPlaneCIDRs...)
		combinedCIDRs = append(combinedCIDRs, cluster.WorkerNodeCIDRs...)
	}
	return combinedCIDRs
}

func ValidateSKIPClusterList(skipClusterList *SKIPClusterList) error {

	if skipClusterList == nil || len(skipClusterList.Clusters) == 0 {
		return errors.New("no SKIPClusterList found")
	}
	if len(skipClusterList.CombinedCIDRS()) == 0 {
		return errors.New("no CIDR ranges in SKIPClusterList")
	}

	// check for empty names, strings and validate that the strings are valid CIDRs
	for _, element := range skipClusterList.Clusters {
		err := checkSKIPClusterFieldsAreNotEmpty(element)
		if err != nil {
			return err
		}
	}

	for _, element := range skipClusterList.CombinedCIDRS() {
		err := checkValidCIDR(element)
		if err != nil {
			return err
		}
	}
	return nil
}

func checkSKIPClusterFieldsAreNotEmpty(cluster *SKIPCluster) error {
	if len(strings.TrimSpace(cluster.Name)) == 0 {
		return errors.New("SKIPCluster name cannot be empty")
	}
	if (len(cluster.WorkerNodeCIDRs) == 0) && (len(cluster.ControlPlaneCIDRs) == 0) {
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
