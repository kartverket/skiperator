package volume

import (
	"strings"

	"github.com/kartverket/skiperator/api/common/podtypes"
	"github.com/kartverket/skiperator/pkg/util"
	"github.com/nais/liberator/pkg/namegen"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/validation"
)

const (
	DefaultDigdiratorMaskinportenMountPath = "/var/run/secrets/skip/maskinporten"
	DefaultDigdiratorIDportenMountPath     = "/var/run/secrets/skip/idporten"

	tmpVolumeName = "tmp"

	// For linking volumes and volume mounts in pods and allowing different volume types to have the same name.
	configMapPrefix             = "cm-"
	secretPrefix                = "sec-"
	emptyDirPrefix              = "ed-"
	persistentVolumeClaimPrefix = "pvc-"
)

// AppendDigdiratorSecret wires a digdirator-issued secret into the container via envFrom
// and mounts it as a file volume. Returns the updated pod volumes and container volume mounts
func AppendDigdiratorSecret(container *corev1.Container, volumeMounts []corev1.VolumeMount, volumes []corev1.Volume, secretName, mountPath string) ([]corev1.Volume, []corev1.VolumeMount) {
	container.EnvFrom = append(container.EnvFrom, corev1.EnvFromSource{
		SecretRef: &corev1.SecretEnvSource{
			LocalObjectReference: corev1.LocalObjectReference{Name: secretName},
		},
	})
	volumeMounts = append(volumeMounts, corev1.VolumeMount{
		Name:      secretName,
		MountPath: mountPath,
		ReadOnly:  true,
	})
	volumes = append(volumes, corev1.Volume{
		Name: secretName,
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName:  secretName,
				Items:       nil,
				DefaultMode: new(int32(420)),
			},
		},
	})
	return volumes, volumeMounts
}

func GetContainerVolumeMounts(filesFrom []podtypes.FilesFrom) []corev1.VolumeMount {
	containerVolumeMounts := []corev1.VolumeMount{
		{
			Name:      tmpVolumeName,
			MountPath: "/tmp",
		},
	}

	for _, file := range filesFrom {
		volumeName := sourceVolumeName(file)
		if volumeName == "" {
			// Skip if no valid volume source is found, should not happen due to kubeAPI CEL validation
			continue
		}
		if len(file.SubPath) > 0 {
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
		tmpVolumeName: {
			Name: tmpVolumeName,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		},
	}

	// Track insertion order to ensure consistent output for testing
	insertionOrder := make([]string, 0)
	insertionOrder = append(insertionOrder, tmpVolumeName)

	for _, file := range filesFrom {
		volume := corev1.Volume{}
		if file.DefaultMode == 0 {
			file.DefaultMode = 420
		}
		if len(file.ConfigMap) > 0 {
			volume = corev1.Volume{
				Name: sourceVolumeName(file),
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
				Name: sourceVolumeName(file),
				VolumeSource: corev1.VolumeSource{
					Secret: &corev1.SecretVolumeSource{
						SecretName:  file.Secret,
						DefaultMode: util.PointTo(int32(file.DefaultMode)),
					},
				},
			}
		} else if len(file.EmptyDir) > 0 {
			if file.EmptyDir == tmpVolumeName {
				continue
			} else {
				volume = corev1.Volume{
					Name: sourceVolumeName(file),
					VolumeSource: corev1.VolumeSource{
						EmptyDir: &corev1.EmptyDirVolumeSource{},
					},
				}
			}
		} else if len(file.PersistentVolumeClaim) > 0 {
			volume = corev1.Volume{
				Name: sourceVolumeName(file),
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

// sourceVolumeName returns the Kubernetes volume name for a FilesFrom source.
// It mirrors the source selection enforced by the CRD: exactly one source field
// should be set. EmptyDir "tmp" maps to the built-in tmp volume.
func sourceVolumeName(file podtypes.FilesFrom) string {
	switch {
	case len(file.ConfigMap) > 0:
		return volumeName(configMapPrefix, file.ConfigMap)
	case len(file.Secret) > 0:
		return volumeName(secretPrefix, file.Secret)
	case len(file.EmptyDir) > 0:
		if file.EmptyDir == tmpVolumeName {
			return tmpVolumeName
		}
		return volumeName(emptyDirPrefix, file.EmptyDir)
	case len(file.PersistentVolumeClaim) > 0:
		return volumeName(persistentVolumeClaimPrefix, file.PersistentVolumeClaim)
	default:
		return ""
	}
}

// volumeName builds a DNS-1123-safe Kubernetes volume name from a typed source
// prefix and the user-provided resource name. Normal names stay readable; names
// that need normalization or truncation get a deterministic hash suffix to avoid
// collisions such as "shared.name" and "shared-name".
func volumeName(prefix, sourceName string) string {
	rawName := prefix + sourceName

	// Kubernetes volume names are DNS-1123 labels, so normalize casing and
	// replace characters like "." that are valid in some source names.
	name := strings.ToLower(rawName)
	name = sanitizeDNS1123Label(name)
	if name == "" {
		name = strings.Trim(prefix, "-")
	}

	// Keep ordinary names unchanged so generated manifests stay readable.
	if name == strings.ToLower(rawName) && len(validation.IsDNS1123Label(name)) == 0 {
		return name
	}

	// Add a hash when normalization or truncation is needed. When sanitization
	// changes the label, hash the raw lowercased source name so different inputs
	// like "shared.name" and "shared_name" do not collapse to the same volume name.
	shortNameInput := name
	if name != strings.ToLower(rawName) {
		shortNameInput = strings.ToLower(rawName)
	}
	shortName, err := namegen.ShortName(shortNameInput, validation.DNS1123LabelMaxLength)
	if err != nil {
		panic(err)
	}
	return sanitizeDNS1123Label(shortName)
}

func sanitizeDNS1123Label(name string) string {
	return strings.Trim(strings.Map(func(r rune) rune {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '-' {
			return r
		}
		return '-'
	}, name), "-")
}
