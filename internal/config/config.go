package config

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/kartverket/skiperator/pkg/util"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	cmName      = "skiperator-config"
	cmNamespace = "skiperator-system"
	cmKey       = "config.json"
)

// SkiperatorConfig holds various configuration options for Skiperator that may differ across
// environments or deployments (public cloud, local development or on-premises).
//
// TODO: Reduce other ways of configuring Skiperator, such as environment variables or flags and use this ConfigMap instead.
type SkiperatorConfig struct {
	TopologyKeys []string `json:"topologyKeys"`
}

var (
	// ActiveConfig holds the currently loaded Skiperator configuration from external sources.
	activeConfig SkiperatorConfig
	mu           sync.Mutex
)

// LoadConfig loads the configuration once at startup
func LoadConfig(ctx context.Context, c client.Client) error {
	mu.Lock()
	defer mu.Unlock()

	cm, err := util.GetConfigMap(c, ctx, types.NamespacedName{Namespace: cmNamespace, Name: cmName})
	if err != nil {
		return fmt.Errorf("failed to load config ConfigMap: %w", err)
	}

	raw, ok := cm.Data[cmKey]
	if !ok {
		return fmt.Errorf("config ConfigMap missing key %q", cmKey)
	}
	if strings.TrimSpace(raw) == "" {
		return fmt.Errorf("config.json is present but empty")
	}

	dec := json.NewDecoder(strings.NewReader(raw))
	dec.DisallowUnknownFields()

	var cfg SkiperatorConfig
	if err := dec.Decode(&cfg); err != nil {
		return fmt.Errorf("failed to unmarshal ConfigMap data: %w", err)
	}

	// Set a default value for TopologyKeys if not set
	if len(cfg.TopologyKeys) == 0 {
		cfg.TopologyKeys = []string{"kubernetes.io/hostname"}
	}

	activeConfig = cfg

	return nil
}

func GetActiveConfig() SkiperatorConfig {
	return activeConfig
}
