package core

// Based on https://kubernetes.io/docs/reference/labels-annotations-taints/

type SkiperatorTopologyKey string

const (
	// Hostname is the value populated by the Kubelet.
	Hostname SkiperatorTopologyKey = "kubernetes.io/hostname"
	// OnPremFailureDomain is populated to the underlying ESXi hostname by the GKE on VMware tooling.
	OnPremFailureDomain SkiperatorTopologyKey = "onprem.gke.io/failure-domain-name"
)
