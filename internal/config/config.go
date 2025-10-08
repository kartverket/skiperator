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
		LeaderElectionNamespace:     "skiperator-system",
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
