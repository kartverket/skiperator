package utils

import (
	"context"
	"fmt"
	vault "github.com/hashicorp/vault/api"
	auth "github.com/hashicorp/vault/api/auth/kubernetes"
)

type Vault struct {
	client *vault.Client
}

func NewVault(ctx context.Context, address, role, token string) (Vault, error) {
	config := vault.DefaultConfig()
	config.Address = address

	client, err := vault.NewClient(config)
	if err != nil {
		err = fmt.Errorf("unable to initialize Vault client: %v", err)
		return Vault{}, err
	}

	k8sAuth, err := auth.NewKubernetesAuth(role, auth.WithServiceAccountToken(token))
	if err != nil {
		err = fmt.Errorf("unable to initialize Kubernetes auth method: %w", err)
		return Vault{}, err
	}

	_, err = client.Auth().Login(ctx, k8sAuth)
	if err != nil {
		err = fmt.Errorf("unable to log in with Kubernetes auth: %w", err)
		return Vault{}, err
	}

	return Vault{client}, nil
}

func (v *Vault) GetSecretJson(ctx context.Context, mountPath, secretPath string) (map[string]any, error) {
	secret, err := v.client.KVv2(mountPath).Get(ctx, secretPath)
	if err != nil {
		err = fmt.Errorf("unable to read secret: %v", err)
		return nil, err
	}

	return secret.Data, nil
}

func (v *Vault) GetSecretString(ctx context.Context, mountPath, secretPath, secretKey string) (string, error) {
	json, err := v.GetSecretJson(ctx, mountPath, secretPath)
	if err != nil {
		return "", err
	}

	value, ok := json[secretKey].(string)
	if !ok {
		err = fmt.Errorf("%s/%s %s is not a string", mountPath, secretPath, secretKey)
		return "", err
	}

	return value, nil
}
