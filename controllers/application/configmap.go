package applicationcontroller

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

var controllerName = "ConfigMap"

func (r *ApplicationReconciler) reconcileConfigMap(ctx context.Context, application *skiperatorv1alpha1.Application) (reconcile.Result, error) {
	r.SetControllerProgressing(ctx, application, controllerName)

	// Is this an error?
	if application.Spec.GCP != nil {
		gcpIdentityConfigMap := corev1.ConfigMap{}

		err := r.GetClient().Get(ctx, types.NamespacedName{Namespace: "skiperator-system", Name: "gcp-identity-config"}, &gcpIdentityConfigMap)
		if errors.IsNotFound(err) {
			r.GetRecorder().Eventf(
				application,
				corev1.EventTypeWarning, "Missing",
				"Cannot find configmap named gcp-identity-config in namespace skiperator-system",
			)
		} else if err != nil {
			r.SetControllerError(ctx, application, controllerName, err)
			return reconcile.Result{}, err
		}

		err = r.setupGCPAuthConfigMap(ctx, gcpIdentityConfigMap, application)
		if err != nil {
			r.SetControllerError(ctx, application, controllerName, err)
			return reconcile.Result{}, err
		}
	}

	err := r.setupSkiperatorConfigMap(ctx, application)

	r.SetControllerFinishedOutcome(ctx, application, controllerName, err)

	return reconcile.Result{}, err

}

func (r *ApplicationReconciler) setupGCPAuthConfigMap(ctx context.Context, gcpIdentityConfigMap corev1.ConfigMap, application *skiperatorv1alpha1.Application) error {

	gcpAuthConfigMapName := application.Name + "-gcp-auth"
	gcpAuthConfigMap := corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Namespace: application.Namespace, Name: gcpAuthConfigMapName}}

	_, err := ctrlutil.CreateOrPatch(ctx, r.GetClient(), &gcpAuthConfigMap, func() error {
		// Set application as owner of the configmap
		err := ctrlutil.SetControllerReference(application, &gcpAuthConfigMap, r.GetScheme())
		if err != nil {
			r.SetControllerError(ctx, application, controllerName, err)
			return err
		}
		r.SetLabelsFromApplication(ctx, &gcpAuthConfigMap, *application)

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
			r.SetControllerError(ctx, application, controllerName, err)
			return err
		}

		gcpAuthConfigMap.Data = map[string]string{
			"config": string(ConfByte),
		}

		return nil
	})

	return err
}

func (r *ApplicationReconciler) setupSkiperatorConfigMap(ctx context.Context, application *skiperatorv1alpha1.Application) error {
	skiperatorConfigMap, err := r.GetSkiperatorConfigMap(ctx, application)

	if errors.IsNotFound(err) {
		r.GetRecorder().Eventf(
			application,
			corev1.EventTypeWarning, "Missing",
			"Cannot find Skiperator ConfigMap, creating",
		)

		err := r.GetClient().Create(ctx, &skiperatorConfigMap)

		if err != nil {
			return err
		}
	}

	mapData := skiperatorConfigMap.Data
	istioSidecarLabelName := "istioSidecarCPURequest"

	// We only want to set the istio CPU Request if it is not already set
	_, present := mapData[istioSidecarLabelName]
	if !present {
		if len(mapData) == 0 {
			mapData = make(map[string]string)
		}
		mapData[istioSidecarLabelName] = getDefaultIstioCPURequestFromEnv(r.Environment)
	}

	skiperatorConfigMap.Data = mapData
	err = r.GetClient().Update(ctx, &skiperatorConfigMap)
	if err != nil {
		r.GetRecorder().Eventf(
			application,
			corev1.EventTypeWarning, "Error",
			"Something went wrong when updating Skiperator ConfigMap: "+err.Error(),
		)
		return err
	}

	return err
}

func (r *ApplicationReconciler) GetSkiperatorConfigMap(ctx context.Context, application *skiperatorv1alpha1.Application) (corev1.ConfigMap, error) {

	skiperatorConfigMapName := application.Name + "-skiperator-config"
	skiperatorConfigMap := corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Namespace: application.Namespace, Name: skiperatorConfigMapName}}

	err := r.GetClient().Get(ctx, types.NamespacedName{Namespace: application.Namespace, Name: skiperatorConfigMapName}, &skiperatorConfigMap)

	return skiperatorConfigMap, err
}

func getDefaultIstioCPURequestFromEnv(env string) string {
	switch env {
	case "prod":
		return "100m"
	case "sandbox", "dev", "test":
		return "10m"
	default:
		// Better to safeguard a high request in case of poor config
		return "100m"
	}

}
