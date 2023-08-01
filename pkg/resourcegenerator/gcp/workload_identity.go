package gcp

import (
	"context"
	"encoding/json"
	"github.com/kartverket/skiperator/pkg/util"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

var (
	CredentialsMountPath = "/var/run/secrets/tokens/gcp-ksa/"
	CredentialsFileName  = "google-application-credentials.json"
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

func GetGoogleServiceAccountCredentialsConfigMap(ctx context.Context, namespace string, name string, gcpServiceAccount string, workloadIdentityConfigMap corev1.ConfigMap) (corev1.ConfigMap, error) {
	logger := log.FromContext(ctx)
	gcpConfigMap := corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: name}}

	credentials := WorkloadIdentityCredentials{
		Type:                           "external_account",
		Audience:                       "identitynamespace:" + workloadIdentityConfigMap.Data["workloadIdentityPool"] + ":" + workloadIdentityConfigMap.Data["identityProvider"],
		ServiceAccountImpersonationUrl: "https://iamcredentials.googleapis.com/v1/projects/-/serviceAccounts/" + gcpServiceAccount + ":generateAccessToken",
		SubjectTokenType:               "urn:ietf:params:oauth:token-type:jwt",
		TokenUrl:                       "https://sts.googleapis.com/v1/token",
		CredentialSource: CredentialSource{
			File: CredentialsMountPath + "token",
		},
	}

	gcpConfigMap.ObjectMeta.Annotations = util.CommonAnnotations

	credentialsBytes, err := json.Marshal(credentials)
	if err != nil {
		logger.Error(err, "could not marshall gcp identity config map")
		return corev1.ConfigMap{}, err
	}

	gcpConfigMap.Data = map[string]string{
		"config": string(credentialsBytes),
	}

	return gcpConfigMap, nil
}

func GetGCPConfigMapName(ownerName string) string {
	return ownerName + "-gcp-auth"
}

func GetGCPEnvVar() corev1.EnvVar {
	return corev1.EnvVar{
		Name:  "GOOGLE_APPLICATION_CREDENTIALS",
		Value: CredentialsMountPath + CredentialsFileName,
	}
}

func GetGCPContainerVolume(workloadIdentityPool string, name string) corev1.Volume {
	twoDaysInSeconds := int64(172800)

	return corev1.Volume{
		Name: "gcp-ksa",
		VolumeSource: corev1.VolumeSource{
			Projected: &corev1.ProjectedVolumeSource{
				DefaultMode: util.PointTo(int32(420)),
				Sources: []corev1.VolumeProjection{
					{
						ServiceAccountToken: &corev1.ServiceAccountTokenProjection{
							Path:              "token",
							Audience:          workloadIdentityPool,
							ExpirationSeconds: &twoDaysInSeconds,
						},
					},
					{
						ConfigMap: &corev1.ConfigMapProjection{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: GetGCPConfigMapName(name),
							},
							Optional: util.PointTo(false),
							Items: []corev1.KeyToPath{
								{
									Key:  "config",
									Path: CredentialsFileName,
								},
							},
						},
					},
				},
			},
		},
	}
}

func GetGCPContainerVolumeMount() corev1.VolumeMount {
	return corev1.VolumeMount{
		Name:      "gcp-ksa",
		MountPath: CredentialsMountPath,
		ReadOnly:  true,
	}
}
