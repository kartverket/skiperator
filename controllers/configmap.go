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
	gcpIdentityConfigMap := corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Namespace: "skiperator-system", Name: "gcpIdentity-config"}}
	_, err := ctrlutil.CreateOrPatch(ctx, r.client, &gcpAuthConfigMap, func() error {
		// Set application as owner of the sidecar
		err := ctrlutil.SetControllerReference(application, &gcpAuthConfigMap, r.scheme)
		if err != nil {
			return err
		}
		
		var ConfStruct Config

		ConfStruct = Config {
			Type: "external_account",
			Audience: "identitynamespace:" + gcpIdentityConfigMap.Data.workloadIdentityPool + ":" + gcpIdentityConfigMap.Data.identityProvider,
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

		gcpAuthConfigMap.Data.Config = string(ConfByte)

		

		return nil
	})
	return reconcile.Result{}, err
}

//kind: ConfigMap
//apiVersion: v1
//metadata:
//  namespace: shipwreck
//  name: my-cloudsdk-config
//data:
//  config: |
//    {
//      "type": "external_account",
//      "audience": "identitynamespace:kubernetes-dev-94b9.svc.id.goog:https://gkehub.googleapis.com/projects/kubernetes-dev-94b9/locations/global/memberships/atkv1-dev",
//      "service_account_impersonation_url": "https://iamcredentials.googleapis.com/v1/projects/-/serviceAccounts/secret-accessor@skip-dev-7d22.iam.gserviceaccount.com:generateAccessToken",
//      "subject_token_type": "urn:ietf:params:oauth:token-type:jwt",
//      "token_url": "https://sts.googleapis.com/v1/token",
//      "credential_source": {
//        "file": "/var/run/secrets/tokens/gcp-ksa/token"
//      }
//    }
//
