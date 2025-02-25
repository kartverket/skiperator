package secret

import (
	"fmt"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/reconciliation"
	"github.com/kartverket/skiperator/pkg/util"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func Generate(r reconciliation.Reconciliation) error {
	ctxLog := r.GetLogger()
	if r.GetType() != reconciliation.ApplicationType {
		return fmt.Errorf("unsupported type %s in auto login secret", r.GetType())
	}
	application, ok := r.GetSKIPObject().(*skiperatorv1alpha1.Application)
	if !ok {
		err := fmt.Errorf("failed to cast resource to application")
		ctxLog.Error(err, "Failed to generate auto login secret")
		return err
	}
	ctxLog.Debug("Attempting to generate secret for envoy filter for application", "application", application.Name)

	autoLoginConfig := r.GetAutoLoginConfig()
	if autoLoginConfig == nil {
		ctxLog.Debug("No auto login config provided for application. Skipping generating secret for envoy filter", "application", application.Name)
		return nil
	}
	envoySecret, err := getEnvoySecret(application.Namespace, autoLoginConfig.ClientSecret)
	if err != nil {
		ctxLog.Error(err, "Failed to generate envoy secret for envoy filter")
		return err
	}
	r.AddResource(envoySecret)
	ctxLog.Debug("Finished generating secret for envoy filter for application", "application", application.Name)
	return nil
}

func getEnvoySecret(namespace string, clientSecret string) (*v1.Secret, error) {
	secretData := map[string][]byte{}

	tokenSecretDataValue, err := getEnvoySecretDataValue("token", clientSecret, "inline_string")
	if err != nil {
		return nil, err
	}
	secretData["token-secret.yaml"] = *tokenSecretDataValue

	hmacSecret, err := util.GenerateHMACSecret(32)
	if err != nil {
		return nil, err
	}
	hmacSecretDataValue, err := getEnvoySecretDataValue("hmac", *hmacSecret, "inline_bytes")
	if err != nil {
		return nil, err
	}
	secretData["hmac-secret.yaml"] = *hmacSecretDataValue

	return &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "auto-login-envoy-secret",
			Namespace: namespace,
		},
		Type: v1.SecretTypeOpaque,
		Data: secretData,
	}, nil
}

func getEnvoySecretDataValue(resourceName string, secret string, secretType string) (*[]byte, error) {
	data := map[string]interface{}{
		"resources": []map[string]interface{}{
			{
				"@type": "type.googleapis.com/envoy.extensions.transport_sockets.tls.v3.Secret",
				"name":  resourceName,
				"generic_secret": map[string]interface{}{
					"secret": map[string]string{
						secretType: secret,
					},
				},
			},
		},
	}
	yamlData, err := yaml.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to mashal yaml: %w", err)
	}
	return &yamlData, nil
}
