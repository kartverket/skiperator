package controllers

import (
	"context"
	"encoding/json"

	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

//+kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;patch;delete

type Config struct {
	Type                           string           `json:"type"`
	Audience                       string           `json:"audience"`
	ServiceAccountImpersonationUrl string           `json:"service_account_impersonation_url"`
	SubjectTokenType               string           `json:"subject_token_type"`
	TokenUrl                       string           `json:"token_url"`
	CredentialSource               CredentialSource `json:"credential_source"`
}
type CredentialSource struct {
	File string `json:"file"`
}

func (r *ApplicationReconciler) reconcileConfigMap(ctx context.Context, application *skiperatorv1alpha1.Application) (reconcile.Result, error) {
	controllerName := "ConfigMap"
	r.ManageControllerStatus(ctx, application, controllerName, skiperatorv1alpha1.PROGRESSING)

	// Is this an error?
	if application.Spec.GCP == nil {
		r.ManageControllerStatus(ctx, application, controllerName, skiperatorv1alpha1.SYNCED)
		return reconcile.Result{}, nil
	}

	gcpIdentityConfigMap := corev1.ConfigMap{}

	err := r.GetClient().Get(ctx, types.NamespacedName{Namespace: "skiperator-system", Name: "gcp-identity-config"}, &gcpIdentityConfigMap)
	if errors.IsNotFound(err) {
		r.GetRecorder().Eventf(
			application,
			corev1.EventTypeWarning, "Missing",
			"Cannot find configmap named gcp-identity-config in namespace skiperator-system",
		)
	} else if err != nil {
		r.ManageControllerStatusError(ctx, application, controllerName, err)
		return reconcile.Result{}, err
	}
	gcpAuthConfigMapName := application.Name + "-gcp-auth"
	gcpAuthConfigMap := corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Namespace: application.Namespace, Name: gcpAuthConfigMapName}}
	_, err = ctrlutil.CreateOrPatch(ctx, r.GetClient(), &gcpAuthConfigMap, func() error {
		// Set application as owner of the configmap
		err := ctrlutil.SetControllerReference(application, &gcpAuthConfigMap, r.GetScheme())
		if err != nil {
			r.ManageControllerStatusError(ctx, application, controllerName, err)
			return err
		}

		if application.Spec.GCP != nil {
			ConfStruct := Config{
				Type:                           "external_account",
				Audience:                       "identitynamespace:" + gcpIdentityConfigMap.Data["workloadIdentityPool"] + ":" + gcpIdentityConfigMap.Data["identityProvider"],
				ServiceAccountImpersonationUrl: "https://iamcredentials.googleapis.com/v1/projects/-/serviceAccounts/" + application.Spec.GCP.Auth.ServiceAccount + ":generateAccessToken",
				SubjectTokenType:               "urn:ietf:params:oauth:token-type:jwt",
				TokenUrl:                       "https://sts.googleapis.com/v1/token",
				CredentialSource: CredentialSource{
					File: "/var/run/secrets/tokens/gcp-ksa/token",
				},
			}

			ConfByte, err := json.Marshal(ConfStruct)
			if err != nil {
				r.ManageControllerStatusError(ctx, application, controllerName, err)
				return err
			}

			gcpAuthConfigMap.Data = map[string]string{
				"config": string(ConfByte),
			}
		}

		return nil
	})

	r.ManageControllerOutcome(ctx, application, controllerName, skiperatorv1alpha1.SYNCED, err)

	return reconcile.Result{}, err

}
