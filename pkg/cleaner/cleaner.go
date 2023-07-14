package cleaner

import (
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

type DeploymentCleaner struct {
	RemovePrefixPredicate func(candidate string) bool
	RemoveEnvVars         []string
	RemoveVolumes         []string
	RemoveInitContainers  []string
}

func (l *DeploymentCleaner) Apply(deployment *v1.Deployment) *v1.Deployment {
	if deployment == nil {
		return nil
	}

	copy := *deployment
	for label := range copy.Labels {
		if l.RemovePrefixPredicate(label) {
			delete(copy.Labels, label)
		}
	}

	for label := range copy.Spec.Template.Labels {
		if l.RemovePrefixPredicate(label) {
			delete(copy.Spec.Template.Labels, label)
		}
	}

	for i := range copy.Spec.Template.Spec.Containers {
		copy.Spec.Template.Spec.Containers[i].Env = removeFromSlice(copy.Spec.Template.Spec.Containers[i].Env, func(e corev1.EnvVar) bool {
			for _, toRemove := range l.RemoveEnvVars {
				if e.Name == toRemove {
					return true
				}
			}
			return false
		})

		copy.Spec.Template.Spec.Containers[i].VolumeMounts = removeFromSlice(copy.Spec.Template.Spec.Containers[i].VolumeMounts, func(e corev1.VolumeMount) bool {
			for _, toRemove := range l.RemoveVolumes {
				if e.Name == toRemove {
					return true
				}
			}
			return false
		})

		copy.Spec.Template.Spec.Containers[i].VolumeMounts = removeFromSlice(copy.Spec.Template.Spec.Containers[i].VolumeMounts, func(e corev1.VolumeMount) bool {
			for _, toRemove := range l.RemoveVolumes {
				if e.Name == toRemove {
					return true
				}
			}
			return false
		})
	}

	copy.Spec.Template.Spec.Volumes = removeFromSlice(copy.Spec.Template.Spec.Volumes, func(e corev1.Volume) bool {
		for _, toRemove := range l.RemoveVolumes {
			if e.Name == toRemove {
				return true
			}
		}
		return false
	})

	copy.Spec.Template.Spec.InitContainers = removeFromSlice(copy.Spec.Template.Spec.InitContainers, func(e corev1.Container) bool {
		for _, toRemove := range l.RemoveInitContainers {
			if e.Name == toRemove {
				return true
			}
		}
		return false
	})

	return &copy
}

func removeFromSlice[S ~[]E, E any](slice S, f func(e E) bool) S {
	var res = *(new(S))
	for _, e := range slice {
		if !f(e) {
			res = append(res, e)
		}
	}

	return res
}
