package controllers

import (
	"context"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"encoding/json"
)

//+kubebuilder:rbac:groups=skiperator.kartverket.no,resources=applications,verbs=get;list;watch
//+kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;patch;delete

type Config struct {
	Type string `json:"type"`
	Audience string `json:"audience"`
	ServiceAccountImpersonationUrl string `json:"service_account_impersonation_url"`
	SubjectTokenType string `json:"subject_token_type"`
	TokenUrl string `json:"token_url"`
	CredentialSource CredentialSource `json:"credential_source"`
}
type CredentialSource struct {
	File string `json:"file"`
}

type ConfigMapReconciler struct {
	client client.Client
	scheme *runtime.Scheme
}

func (r *ConfigMapReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.client = mgr.GetClient()
	r.scheme = mgr.GetScheme()

	return newControllerManagedBy[*skiperatorv1alpha1.Application](mgr).
		For(&skiperatorv1alpha1.Application{}).
		Owns(&corev1.ConfigMap{}).
		Complete(r)
}

func (r *ConfigMapReconciler) Reconcile(ctx context.Context, application *skiperatorv1alpha1.Application) (reconcile.Result, error) {
	application.FillDefaults()

	gcpAuthConfigMapName := application.Name + "-gcp-auth"

	gcpAuthConfigMap := corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Namespace: application.Namespace, Name: gcpAuthConfigMapName}}
	gcpIdentityConfigMap := corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Namespace: "skiperator-system", Name: "gcpidentityconfig"}}
	
	
	_, err := ctrlutil.CreateOrPatch(ctx, r.client, &gcpAuthConfigMap, func() error {
		// Set application as owner of the configmap
		err := ctrlutil.SetControllerReference(application, &gcpAuthConfigMap, r.scheme)
		if err != nil {
			return err
		}

		ConfStruct := Config {
			Type: "external_account",
			Audience: "identitynamespace:" + gcpIdentityConfigMap.Data["workloadIdentityPool"] + ":" + gcpIdentityConfigMap.Data["identityProvider"],
			ServiceAccountImpersonationUrl: "https://iamcredentials.googleapis.com/v1/projects/-/serviceAccounts/" + application.Spec.GCP.Auth.ServiceAccount + "@" + application.Spec.GCP.Auth.Project + ".iam.gserviceaccount.com:generateAccessToken",
			SubjectTokenType: "urn:ietf:params:oauth:token-type:jwt",
			TokenUrl: "https://sts.googleapis.com/v1/token",
			CredentialSource: CredentialSource{
				File: "/var/run/secrets/tokens/gcp-ksa/token",
			} ,
		}

		ConfByte, err := json.Marshal(ConfStruct)
		if err != nil {
			return err
		}
		ConMap := map[string]string{}
		json.Unmarshal(ConfByte, &ConMap) 
		
		
		gcpAuthConfigMap.Data = ConMap
		
		return nil
	})

	return reconcile.Result{}, err
}

//ConfStruct := Config {
//	Type: "external_account",
//	Audience: "identitynamespace:" + string(gcpIdentityConfigMap.Data.workloadIdentityPool) + ":" + string(gcpIdentityConfigMap.Data.identityProvider),
//	ServiceAccountImpersonationUrl: "https://iamcredentials.googleapis.com/v1/projects/-/serviceAccounts/" + application.Spec.GCP.Auth.ServiceAccount + "@" + application.Spec.GCP.Auth.Project + ".iam.gserviceaccount.com:generateAccessToken",
//	SubjectTokenType: "urn:ietf:params:oauth:token-type:jwt",
//	TokenUrl: "https://sts.googleapis.com/v1/token",
//	CredentialSource: CredentialSource{
//		File: "/var/run/secrets/tokens/gcp-ksa/token",
//	} ,
//}
