package imagepullsecret

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/kartverket/skiperator/pkg/reconciliation"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ImagePullSecret struct {
	payload []byte
}

type RegistryCredentialSecret struct {
	Registry  string        `json:"registry"`  // URL to an image registry, ie https://index.docker.io/v1/ or https://ghcr.io
	Secret    corev1.Secret `json:"secret"`    // The secret containing the auth to the image registry
	SecretKey string        `json:"secretKey"` // The secret key containing the auth to the image registry
}

func NewImagePullSecret(registries ...RegistryCredentialSecret) (*ImagePullSecret, error) {
	cfg := dockerConfigJson{}
	cfg.Auths = make(map[string]dockerConfigAuth, len(registries))

	for _, r := range registries {
		auth := strings.TrimSuffix(string(r.Secret.Data[r.SecretKey]), "\n")
		cfg.Auths[r.Registry] = dockerConfigAuth{
			Auth: auth,
		}
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	err := enc.Encode(cfg)
	if err != nil {
		return nil, err
	}

	return &ImagePullSecret{payload: buf.Bytes()}, nil
}

func (ips *ImagePullSecret) Generate(r reconciliation.Reconciliation) error {
	if r.GetType() != reconciliation.NamespaceType {
		return fmt.Errorf("image pull secret only supports namespace type")
	}
	// SKIPObject here is a namespace, so thats why we use GetName, not GetNamespace.
	// Should NOT be called from any other controller than namespace-controller
	secret := corev1.Secret{ObjectMeta: metav1.ObjectMeta{Namespace: r.GetSKIPObject().GetName(), Name: "github-auth"}}
	secret.Type = corev1.SecretTypeDockerConfigJson

	secret.Data = make(map[string][]byte, 1)
	secret.Data[".dockerconfigjson"] = ips.payload

	r.AddResource(&secret)
	return nil
}

// IsImagePullSecret filters for secrets named github-auth
func IsImagePullSecret(secret *corev1.Secret) bool {
	return secret.Name == "github-auth"
}

type dockerConfigJson struct {
	Auths map[string]dockerConfigAuth `json:"auths"`
}

type dockerConfigAuth struct {
	Auth string `json:"auth"`
}
