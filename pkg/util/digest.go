package util

import (
	"context"
	"encoding/json"

	"github.com/go-logr/logr"
	"github.com/google/k8s-digester/pkg/resolve"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func ResolveImageTags(ctx context.Context, log logr.Logger, config *rest.Config, obj client.Object) error {
	n, err := parseManifest(obj)
	if err != nil {
		return err
	}

	if err = resolve.ImageTags(ctx, log, config, n, []string{}); err != nil {
		return err
	}

	b, _ := n.MarshalJSON()
	if err = json.Unmarshal(b, obj); err != nil {
		return err
	}

	return nil
}

func parseManifest(obj client.Object) (*yaml.RNode, error) {
	m, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}

	return yaml.ConvertJSONToYamlNode(string(m))
}
