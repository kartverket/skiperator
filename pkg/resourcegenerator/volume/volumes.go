package volume

import (
	"github.com/kartverket/skiperator/api/v1alpha1/podtypes"
	"github.com/kartverket/skiperator/pkg/util"
	corev1 "k8s.io/api/core/v1"
)

func GetContainerVolumeMounts(filesFrom []podtypes.FilesFrom) []corev1.VolumeMount {
	containerVolumeMounts := []corev1.VolumeMount{
		{
			Name:      "tmp",
			MountPath: "/tmp",
		},
	}

	for _, file := range filesFrom {

		volumeName := ""
		if len(file.ConfigMap) > 0 {
			volumeName = file.ConfigMap
		} else if len(file.Secret) > 0 {
			volumeName = file.Secret
		} else if len(file.EmptyDir) > 0 {
			volumeName = file.EmptyDir
		} else if len(file.PersistentVolumeClaim) > 0 {
			volumeName = file.PersistentVolumeClaim
		}
		if volumeName == "" {
			// Skip if no valid volume source is found, should not happen due to kubeAPI CEL validation
			continue
		}
		if file.SubPath != "" {
			containerVolumeMounts = append(containerVolumeMounts, corev1.VolumeMount{
				Name:      volumeName,
				MountPath: file.MountPath,
				SubPath:   file.SubPath,
			})
		} else {
			containerVolumeMounts = append(containerVolumeMounts, corev1.VolumeMount{
				Name:      volumeName,
				MountPath: file.MountPath,
			})
		}
	}

	return containerVolumeMounts
}

func GetPodVolumes(filesFrom []podtypes.FilesFrom) []corev1.Volume {

	// Use a map to avoid duplicates
	podVolumesMap := map[string]corev1.Volume{
		"tmp": {
			Name: "tmp",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		},
	}

	// Track insertion order to ensure consistent output for testing
	insertionOrder := make([]string, 0)
	insertionOrder = append(insertionOrder, "tmp")

	for _, file := range filesFrom {
		volume := corev1.Volume{}
		if file.DefaultMode == 0 {
			file.DefaultMode = 420
		}
		if len(file.ConfigMap) > 0 {
			volume = corev1.Volume{
				Name: file.ConfigMap,
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: file.ConfigMap,
						},
						DefaultMode: util.PointTo(int32(file.DefaultMode)),
					},
				},
			}
		} else if len(file.Secret) > 0 {
			volume = corev1.Volume{
				Name: file.Secret,
				VolumeSource: corev1.VolumeSource{
					Secret: &corev1.SecretVolumeSource{
						SecretName:  file.Secret,
						DefaultMode: util.PointTo(int32(file.DefaultMode)),
					},
				},
			}
		} else if len(file.EmptyDir) > 0 {
			volume = corev1.Volume{
				Name: file.EmptyDir,
				VolumeSource: corev1.VolumeSource{
					EmptyDir: &corev1.EmptyDirVolumeSource{},
				},
			}
		} else if len(file.PersistentVolumeClaim) > 0 {
			volume = corev1.Volume{
				Name: file.PersistentVolumeClaim,
				VolumeSource: corev1.VolumeSource{
					PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
						ClaimName: file.PersistentVolumeClaim,
					},
				},
			}
		}

		if _, exists := podVolumesMap[volume.Name]; !exists {
			podVolumesMap[volume.Name] = volume
			insertionOrder = append(insertionOrder, volume.Name)
		}
	}

	podVolumes := make([]corev1.Volume, 0, len(insertionOrder))
	for _, key := range insertionOrder {
		podVolumes = append(podVolumes, podVolumesMap[key])
	}

	return podVolumes
}
