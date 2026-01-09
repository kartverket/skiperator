package gcp

import (
	"fmt"
	"github.com/kartverket/skiperator/v2/pkg/util"
	corev1 "k8s.io/api/core/v1"
)

// TODO move to a more suitable pkg
var (
	CredentialsMountPath = "/var/run/secrets/tokens/gcp-ksa"
	CredentialsFileName  = "google-application-credentials.json"

	ServiceAccountTokenExpiration = int64(60 * 60 * 24 * 2) // Two days
)

func GetGCPConfigMapName(ownerName string) string {
	return ownerName + "-gcp-auth"
}

func GetGCPEnvVar() corev1.EnvVar {
	return corev1.EnvVar{
		Name:  "GOOGLE_APPLICATION_CREDENTIALS",
		Value: fmt.Sprintf("%v/%v", CredentialsMountPath, CredentialsFileName),
	}
}

func GetGCPContainerVolume(workloadIdentityPool string, name string) corev1.Volume {
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
							ExpirationSeconds: &ServiceAccountTokenExpiration,
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
