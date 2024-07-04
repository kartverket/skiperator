package auth

import (
	"context"
	"encoding/json"
	"fmt"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/log"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/gcp"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/resourceutils"
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

func Generate(ctx context.Context, application *skiperatorv1alpha1.Application, gcpIdentityConfigMap *corev1.ConfigMap) (*corev1.ConfigMap, error) {
	ctxLog := log.FromContext(ctx)
	ctxLog.Debug("Generating configmap for application", application.Name)

	gcpAuthConfigMapName := gcp.GetGCPConfigMapName(application.Name)
	gcpConfigMap := corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Namespace: application.Namespace, Name: gcpAuthConfigMapName}}

	credentials := WorkloadIdentityCredentials{
		Type:                           "external_account",
		Audience:                       "identitynamespace:" + gcpIdentityConfigMap.Data["workloadIdentityPool"] + ":" + gcpIdentityConfigMap.Data["identityProvider"],
		ServiceAccountImpersonationUrl: "https://iamcredentials.googleapis.com/v1/projects/-/serviceAccounts/" + application.Spec.GCP.Auth.ServiceAccount + ":generateAccessToken",
		SubjectTokenType:               "urn:ietf:params:oauth:token-type:jwt",
		TokenUrl:                       "https://sts.googleapis.com/v1/token",
		CredentialSource: CredentialSource{
			File: fmt.Sprintf("%v/token", CredentialsMountPath),
		},
	}

	credentialsBytes, err := json.Marshal(credentials)
	if err != nil {
		ctxLog.Error(err, "could not marshall gcp identity config map")
		return nil, err
	}

	gcpConfigMap.Data = map[string]string{
		"config": string(credentialsBytes),
	}

	resourceutils.SetCommonAnnotations(&gcpConfigMap)
	resourceutils.SetApplicationLabels(&gcpConfigMap, application)

	return &gcpConfigMap, nil
}
