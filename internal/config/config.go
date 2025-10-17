package config

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/kartverket/skiperator/pkg/util"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	cmName      = "skiperator-config"
	cmNamespace = "skiperator-system"
	cmKey       = "config.json"
)

type RegistryCredentialPair struct {
	Registry string `json:"registry"`
	Token    string `json:"token"`
}

// SkiperatorConfig holds various configuration options for Skiperator that may differ across
// environments or deployments (public cloud, local development or on-premises).
// TopologyKeys - what node labels to set when configuring pod topology spread constraints, see
// https://kubernetes.io/docs/concepts/scheduling-eviction/topology-spread-constraints/
// LeaderElection - whether to enable leader election, must be set to true if number of Skiperator replicas > 1
// LeaderElectionNamespace - namespace for leader election
// ConcurrentReconciles - Number of concurrent reconciles that Skiperator will perform.
// May incur performance degradation if set too high or too low.
// IsDeployment - set to true if deploying to a cluster, else set to false. Prevents running local testing against
// actual Kubernetes clusters
// LogLevel - permitted values: info, debug, error
// EnableProfiling - enables the use of pprof to visualize and analyze profiling data of Skiperator
// RegistryCredentials - list of URLS and access tokens for container registries that will be inserted into all
// Skiperator-managed application namespaces
// ClusterCIDRExclusionEnabled - set to true to prevent Skiperator-managed applications from reaching certain CIDR ranges
// like cluster nodes, control plane etc.
// ClusterCIDRMap - map of the CIDR ranges to block traffic from Skiperator-managed application namespaces
// EnableLocallyBuiltImages - whether to enable Skiperator to allow the use of locally built container images
// for development purposes
// GCPIdentityProvider - Provider for Workload Identity Federation (WIF)
// GCPWorkloadIdentityPool - Identity pool for Workload Identity Federation (WIF)
type SkiperatorConfig struct {
	TopologyKeys                []string                 `json:"topologyKeys,omitempty"`
	LeaderElection              bool                     `json:"leaderElection,omitempty"`
	LeaderElectionNamespace     string                   `json:"leaderElectionNamespace,omitempty"`
	ConcurrentReconciles        int                      `json:"concurrentReconciles,omitempty"`
	IsDeployment                bool                     `json:"isDeployment,omitempty"`
	LogLevel                    string                   `json:"logLevel,omitempty"`
	EnableProfiling             bool                     `json:"enableProfiling,omitempty"`
	RegistryCredentials         []RegistryCredentialPair `json:"registryCredentials,omitempty"`
	ClusterCIDRExclusionEnabled bool                     `json:"clusterCIDRExclusionEnabled,omitempty"`
	ClusterCIDRMap              SKIPClusterList          `json:"clusterCIDRMap,omitempty"`
	EnableLocallyBuiltImages    bool                     `json:"enableLocallyBuiltImages,omitempty"`
	GCPIdentityProvider         string                   `json:"gcpIdentityProvider,omitempty"`
	GCPWorkloadIdentityPool     string                   `json:"gcpWorkloadIdentityPool,omitempty"`
}

var (
	// ActiveConfig holds the currently loaded Skiperator configuration from external sources.
	activeConfig SkiperatorConfig
	mu           sync.Mutex
)

func GetActiveConfig() SkiperatorConfig {
	return activeConfig
}

// LoadConfig loads the configuration once at startup
func LoadConfig(ctx context.Context, c client.Client) error {
	mu.Lock()
	defer mu.Unlock()

	cm, err := util.GetConfigMap(c, ctx, types.NamespacedName{Namespace: cmNamespace, Name: cmName})
	if err != nil {
		return fmt.Errorf("failed to load config ConfigMap: %w", err)
	}

	return parseConfig(&cm)
}

func parseConfig(cm *corev1.ConfigMap) error {
	raw, ok := cm.Data[cmKey]
	if !ok {
		return fmt.Errorf("config ConfigMap missing key %q", cmKey)
	}
	if strings.TrimSpace(raw) == "" {
		return fmt.Errorf("config.json is present but empty")
	}

	dec := json.NewDecoder(strings.NewReader(raw))
	dec.DisallowUnknownFields()

	// Default values
	var cfg = SkiperatorConfig{
		TopologyKeys:                []string{"kubernetes.io/hostname"},
		LeaderElection:              false,
		LeaderElectionNamespace:     cmNamespace,
		ConcurrentReconciles:        1,
		IsDeployment:                false,
		LogLevel:                    "info",
		RegistryCredentials:         []RegistryCredentialPair{},
		ClusterCIDRExclusionEnabled: false,
		ClusterCIDRMap:              SKIPClusterList{},
		EnableProfiling:             false,
		EnableLocallyBuiltImages:    false,
		GCPIdentityProvider:         "",
		GCPWorkloadIdentityPool:     "",
	}

	if err := dec.Decode(&cfg); err != nil {
		return fmt.Errorf("failed to unmarshal ConfigMap data: %w", err)
	}

	activeConfig = cfg

	return nil
}
