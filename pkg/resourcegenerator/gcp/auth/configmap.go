package auth

import (
	"encoding/json"
	"fmt"
	"github.com/kartverket/skiperator/pkg/reconciliation"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/gcp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	CredentialsMountPath = "/var/run/secrets/tokens/gcp-ksa"
)

type WorkloadIdentityCredentials struct {
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

func Generate(r reconciliation.Reconciliation) error {
	ctxLog := r.GetLogger()

	if r.GetType() == reconciliation.ApplicationType || r.GetType() == reconciliation.JobType {
		return getConfigMap(r)
	} else {
		err := fmt.Errorf("unsupported type %s in gcp configmap", r.GetType())
		ctxLog.Error(err, "Failed to generate gcp configmap")
		return err
	}
}

func getConfigMap(r reconciliation.Reconciliation) error {
	commonSpec := r.GetSKIPObject().GetCommonSpec()
	if commonSpec.GCP == nil || commonSpec.GCP.Auth == nil || commonSpec.GCP.Auth.ServiceAccount == "" {
		return nil
	}

	ctxLog := r.GetLogger()
	ctxLog.Debug("Generating gcp configmap", "type", r.GetType())

	object := r.GetSKIPObject()
	gcpAuthConfigMapName := gcp.GetGCPConfigMapName(object.GetName())
	gcpConfigMap := corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Namespace: object.GetNamespace(), Name: gcpAuthConfigMapName}}

	credentials := WorkloadIdentityCredentials{
		Type:                           "external_account",
		Audience:                       "identitynamespace:" + r.GetSkiperatorConfig().GCPWorkloadIdentityPool + ":" + r.GetSkiperatorConfig().GCPIdentityProvider,
		ServiceAccountImpersonationUrl: "https://iamcredentials.googleapis.com/v1/projects/-/serviceAccounts/" + r.GetSKIPObject().GetCommonSpec().GCP.Auth.ServiceAccount + ":generateAccessToken",
		SubjectTokenType:               "urn:ietf:params:oauth:token-type:jwt",
		TokenUrl:                       "https://sts.googleapis.com/v1/token",
		CredentialSource: CredentialSource{
			File: fmt.Sprintf("%v/token", CredentialsMountPath),
		},
	}

	credentialsBytes, err := json.Marshal(credentials)
	if err != nil {
		ctxLog.Error(err, "could not marshall gcp identity config map")
		return err
	}

	gcpConfigMap.Data = map[string]string{
		"config": string(credentialsBytes),
	}
	r.AddResource(&gcpConfigMap)

	ctxLog.Debug("Finished generating configmap", "type", r.GetType(), "name", object.GetName())
	return nil
}
