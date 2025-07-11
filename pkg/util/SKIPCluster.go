package util

type SKIPCluster struct {
	Name              string   `yaml:"name"`
	ControlPlaneCIDRs []string `yaml:"controlPlaneCIDRs"`
	WorkerNodeCIDRs   []string `yaml:"workerNodeCIDRs"`
}

type SKIPClusterList struct {
	Clusters []*SKIPCluster `yaml:"clusters"`
}

func ClusterList(clusters ...*SKIPCluster) *SKIPClusterList {
	return &SKIPClusterList{}
}

func (c *SKIPClusterList) CombinedCIDRS() []string {
	var combinedCIDRs []string
	for _, cluster := range c.Clusters {
		combinedCIDRs = append(combinedCIDRs, cluster.ControlPlaneCIDRs...)
		combinedCIDRs = append(combinedCIDRs, cluster.WorkerNodeCIDRs...)
	}
	return combinedCIDRs
}
