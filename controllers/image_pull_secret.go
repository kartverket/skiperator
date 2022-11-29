package controllers

import (
	"bytes"
	"context"
	"encoding/json"

	util "github.com/kartverket/skiperator/pkg/util"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

//+kubebuilder:rbac:groups=core,resources=namespaces,verbs=get;list;watch
//+kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete

type ImagePullSecretReconciler struct {
	client   client.Client
	scheme   *runtime.Scheme
	Registry string
	Token    string
}

func (r *ImagePullSecretReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.client = mgr.GetClient()
	r.scheme = mgr.GetScheme()

	return newControllerManagedBy[*corev1.Namespace](mgr).
		For(&corev1.Namespace{}, builder.WithPredicates(
			matchesPredicate[*corev1.Namespace](util.IsNotExcludedNamespace),
		)).
		Owns(&corev1.Secret{}, builder.WithPredicates(
			matchesPredicate[*corev1.Secret](isImagePullSecret),
		)).
		Complete(r)
}

func (r *ImagePullSecretReconciler) Reconcile(ctx context.Context, namespace *corev1.Namespace) (reconcile.Result, error) {
	secret := corev1.Secret{ObjectMeta: metav1.ObjectMeta{Namespace: namespace.Name, Name: "github-auth"}}
	_, err := ctrlutil.CreateOrPatch(ctx, r.client, &secret, func() error {
		// Set namespace as owner of the sidecar
		err := ctrlutil.SetControllerReference(namespace, &secret, r.scheme)
		if err != nil {
			return err
		}

		secret.Type = corev1.SecretTypeDockerConfigJson

		cfg := dockerConfigJson{}
		cfg.Auths = make(map[string]dockerConfigAuth, 1)
		auth := dockerConfigAuth{}
		auth.Auth = r.Token
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
