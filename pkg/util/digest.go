package util

import (
	"context"
	"encoding/json"
	"github.com/go-logr/logr"
	"github.com/google/k8s-digester/pkg/resolve"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func ResolveImageTags(ctx context.Context, log logr.Logger, config *rest.Config, deployment *appsv1.Deployment) (*appsv1.Deployment, error) {
	n, err := parseManifest(deployment)
	if err != nil {
		return nil, err
	}

	if err = resolve.ImageTags(ctx, log, config, n); err != nil {
		return nil, err
	}

	b, _ := n.MarshalJSON()
	var res appsv1.Deployment
	err = json.Unmarshal(b, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

func parseManifest(deployment *appsv1.Deployment) (*yaml.RNode, error) {
	m, err := json.Marshal(*deployment)
	if err != nil {
		return nil, err
	}

	return yaml.ConvertJSONToYamlNode(string(m[:]))
}
