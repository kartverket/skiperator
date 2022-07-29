package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"strings"
)

//+kubebuilder:rbac:groups=skiperator.kartverket.no,resources=applications,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=authentication.k8s.io,resources=tokenreviews,verbs=create

type RegistrySecretPre124Reconciler struct {
	client   client.Client
	scheme   *runtime.Scheme
	recorder record.EventRecorder

	VaultAddress string
}

func (r *RegistrySecretPre124Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.client = mgr.GetClient()
	r.scheme = mgr.GetScheme()
	r.recorder = mgr.GetEventRecorderFor("registry-secret-controller")

	return ctrl.NewControllerManagedBy(mgr).
		For(&skiperatorv1alpha1.Application{}).
		Owns(&corev1.Secret{}).
		Complete(r)
}

func (r *RegistrySecretPre124Reconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	// Fetch application and fill defaults
	application := skiperatorv1alpha1.Application{}
	err := r.client.Get(ctx, req.NamespacedName, &application)
	if err != nil {
		err = client.IgnoreNotFound(err)
		return reconcile.Result{}, err
	}
	application.FillDefaults()

	secret := corev1.Secret{ObjectMeta: metav1.ObjectMeta{Namespace: req.Namespace, Name: req.Name + "-registry"}}

	if application.Spec.Registry == nil {
		err = r.client.Delete(ctx, &secret)
		err = client.IgnoreNotFound(err)
		return reconcile.Result{}, err
	}

	tokenSecret := corev1.Secret{}
	err = r.client.Get(ctx, types.NamespacedName{Namespace: req.Namespace, Name: req.Name + "-token"}, &tokenSecret)
	if err != nil {
		return reconcile.Result{}, err
	}

	var token string
	if application.Spec.Registry != nil {
		vault, err := utils.NewVault(
			ctx,
			r.VaultAddress,
			application.Name,
			string(tokenSecret.Data["token"]),
		)
		if err != nil {
			r.recorder.Eventf(
				&application, corev1.EventTypeWarning, "Vault",
				"Failed to log into Vault with role %s", application.Name,
			)
			return reconcile.Result{Requeue: true}, nil
		}

		token, err = vault.GetSecretString(
			ctx,
			application.Spec.Registry.MountPath,
			application.Spec.Registry.SecretPath,
			application.Spec.Registry.SecretKey,
		)
		if err != nil {
			r.recorder.Eventf(
				&application, corev1.EventTypeWarning, "Vault",
				"Failed getting %s in %s/%s from Vault",
				application.Spec.Registry.SecretKey, application.Spec.Registry.MountPath, application.Spec.Registry.SecretPath,
			)
			return reconcile.Result{Requeue: true}, nil
		}
	}

	_, err = ctrlutil.CreateOrPatch(ctx, r.client, &secret, func() error {
		// Set application as owner of the sidecar
		err = ctrlutil.SetControllerReference(&application, &secret, r.scheme)
		if err != nil {
			return err
		}

		secret.Type = corev1.SecretTypeDockerConfigJson

		var registry string
		// image format: [[registry[:port]]/repo/repo/]image:tag
		image := strings.SplitN(application.Spec.Image, "/", 3)
		if len(image) == 3 {
			registry = image[0]
		} else {
			registry = "https://index.docker.io/v1/"
		}

		cfg := dockerConfigJson{}
		cfg.Auths = make(map[string]dockerConfigAuth, 1)
		auth := dockerConfigAuth{}
		auth.Auth = token
		cfg.Auths[registry] = auth

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
