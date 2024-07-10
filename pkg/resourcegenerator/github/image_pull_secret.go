package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/kartverket/skiperator/pkg/reconciliation"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TODO figure out something smart so we only need to supply the reconciliation object
func Generate(r reconciliation.Reconciliation, token string, registry string) error {
	if r.GetType() != reconciliation.NamespaceType {
		return fmt.Errorf("image pull secret only supports namespace type")
	}
	secret := corev1.Secret{ObjectMeta: metav1.ObjectMeta{Namespace: r.GetReconciliationObject().GetName(), Name: "github-auth"}}

	secret.Type = corev1.SecretTypeDockerConfigJson

	cfg := dockerConfigJson{}
	cfg.Auths = make(map[string]dockerConfigAuth, 1)
	auth := dockerConfigAuth{}
	auth.Auth = token
	cfg.Auths[registry] = auth

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	err := enc.Encode(cfg)
	if err != nil {
		return err
	}

	secret.Data = make(map[string][]byte, 1)
	secret.Data[".dockerconfigjson"] = buf.Bytes()

	return nil
}

// Filter for secrets named github-auth
func IsImagePullSecret(secret *corev1.Secret) bool {
	return secret.Name == "github-auth"
}

type dockerConfigJson struct {
	Auths map[string]dockerConfigAuth `json:"auths"`
}

type dockerConfigAuth struct {
	Auth string `json:"auth"`
}
