package controllers

import (
	"bytes"
	"context"
	"encoding/json"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// TODO Handle controller status when general status handler exists

func (r *NamespaceReconciler) reconcileImagePullSecret(ctx context.Context, namespace *corev1.Namespace) (reconcile.Result, error) {
	secret := corev1.Secret{ObjectMeta: metav1.ObjectMeta{Namespace: namespace.Name, Name: "github-auth"}}
	_, err := ctrlutil.CreateOrPatch(ctx, r.GetClient(), &secret, func() error {
		// Set namespace as owner of the sidecar
		err := ctrlutil.SetControllerReference(namespace, &secret, r.GetScheme())
		if err != nil {
			return err
		}

		secret.Type = corev1.SecretTypeDockerConfigJson

		cfg := dockerConfigJson{}
		cfg.Auths = make(map[string]dockerConfigAuth, 1)
		auth := dockerConfigAuth{}
		auth.Auth = r.Registry
		cfg.Auths[r.Registry] = auth

		var buf bytes.Buffer
		enc := json.NewEncoder(&buf)
		err = enc.Encode(cfg)
		if err != nil {
			return err
		}

		secret.Data = make(map[string][]byte, 1)
		secret.Data[".dockerconfigjson"] = buf.Bytes()

		return nil
	})
	return reconcile.Result{}, err
}

// Filter for secrets named github-auth
func isImagePullSecret(secret *corev1.Secret) bool {
	return secret.Name == "github-auth"
}

type dockerConfigJson struct {
	Auths map[string]dockerConfigAuth `json:"auths"`
}

type dockerConfigAuth struct {
	Auth string `json:"auth"`
}
