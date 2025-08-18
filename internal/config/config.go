package config

import (
	"context"
	"fmt"
	"strconv"

	"github.com/kartverket/skiperator/pkg/util"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var IsCloudCluster bool

// LoadConfig loads the configuration once at startup
func LoadConfig(client client.Client) error {
	configMapNamespaced := types.NamespacedName{
		Namespace: "skiperator-system",
		Name:      "skiperator-config",
	}

	configMap, err := util.GetConfigMap(client, context.Background(), configMapNamespaced)
	if err != nil {
		return fmt.Errorf("failed to load config ConfigMap: %w", err)
	}

	isCloudStr, exists := configMap.Data["isCloudCluster"]
	if !exists {
		return fmt.Errorf("isCloudCluster field not found in ConfigMap")
	}
	isCloud, err := strconv.ParseBool(isCloudStr)
	if err != nil {
		return fmt.Errorf("invalid boolean value for isCloudCluster: %q: %w", isCloudStr, err)
	}
	IsCloudCluster = isCloud

	return nil
}
