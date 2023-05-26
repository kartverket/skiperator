package pod

import (
	"fmt"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/api/v1alpha1/podtypes"
	"github.com/kartverket/skiperator/pkg/util"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func CreatePodSpec(container corev1.Container, volumes []corev1.Volume, serviceAccountName string, priority string, policy corev1.RestartPolicy) corev1.PodSpec {
	return corev1.PodSpec{
		Volumes: volumes,
		Containers: []corev1.Container{
			container,
		},
		ServiceAccountName: serviceAccountName,
		SecurityContext: &corev1.PodSecurityContext{
			SupplementalGroups: []int64{util.SkiperatorUser},
			FSGroup:            util.PointTo(util.SkiperatorUser),
			SeccompProfile: &corev1.SeccompProfile{
				Type: corev1.SeccompProfileTypeRuntimeDefault,
			},
		},
		ImagePullSecrets:  []corev1.LocalObjectReference{{Name: "github-auth"}},
		PriorityClassName: fmt.Sprintf("skip-%s", priority),
		RestartPolicy:     policy,
	}

}

func CreateApplicationContainer(application *skiperatorv1alpha1.Application) corev1.Container {
	return corev1.Container{
		Name:            application.Name,
		Image:           application.Spec.Image,
		ImagePullPolicy: corev1.PullAlways,
		Command:         application.Spec.Command,
		SecurityContext: &corev1.SecurityContext{
			Privileged:               util.PointTo(false),
			AllowPrivilegeEscalation: util.PointTo(false),
			ReadOnlyRootFilesystem:   util.PointTo(true),
			RunAsUser:                util.PointTo(util.SkiperatorUser),
			RunAsGroup:               util.PointTo(util.SkiperatorUser),
		},
		Ports:          getContainerPorts(application),
		EnvFrom:        getEnvFrom(application.Spec.EnvFrom),
		Resources:      getResourceRequirements(application.Spec.Resources),
		Env:            application.Spec.Env,
		ReadinessProbe: getProbe(application.Spec.Readiness),
		LivenessProbe:  getProbe(application.Spec.Liveness),
		StartupProbe:   getProbe(application.Spec.Startup),
	}
}

func CreateJobContainer(job *skiperatorv1alpha1.SKIPJob) corev1.Container {
	return corev1.Container{
		Name:            job.Name + "-job",
		Image:           job.Spec.Container.Image,
		ImagePullPolicy: corev1.PullAlways,
		Command:         job.Spec.Container.Command,
		SecurityContext: &corev1.SecurityContext{
			Privileged:               util.PointTo(false),
			AllowPrivilegeEscalation: util.PointTo(false),
			ReadOnlyRootFilesystem:   util.PointTo(true),
			RunAsUser:                util.PointTo(util.SkiperatorUser),
			RunAsGroup:               util.PointTo(util.SkiperatorUser),
		},
		EnvFrom:        getEnvFrom(job.Spec.Container.EnvFrom),
		Resources:      getResourceRequirements(job.Spec.Container.Resources),
		Env:            job.Spec.Container.Env,
		ReadinessProbe: getProbe(job.Spec.Container.Readiness),
		LivenessProbe:  getProbe(job.Spec.Container.Liveness),
		StartupProbe:   getProbe(job.Spec.Container.Startup),
	}
}

func getProbe(appProbe *podtypes.Probe) *corev1.Probe {
	if appProbe != nil {
		probe := corev1.Probe{
			InitialDelaySeconds: int32(appProbe.InitialDelay),
			TimeoutSeconds:      int32(appProbe.Timeout),
			FailureThreshold:    int32(appProbe.FailureThreshold),
			ProbeHandler: corev1.ProbeHandler{
				HTTPGet: &corev1.HTTPGetAction{
					Path: appProbe.Path,
					Port: intstr.FromInt(int(appProbe.Port)),
				},
			},
		}

		return &probe
	}

	return nil
}

func getResourceRequirements(resources *podtypes.ResourceRequirements) corev1.ResourceRequirements {
	if resources == nil {
		return corev1.ResourceRequirements{}
	}

	return corev1.ResourceRequirements{
		Limits:   (*resources).Limits,
		Requests: (*resources).Requests,
	}
}

func getEnvFrom(envFromApplication []podtypes.EnvFrom) []corev1.EnvFromSource {
	var envFromSource []corev1.EnvFromSource

	for _, env := range envFromApplication {
		if len(env.ConfigMap) > 0 {
			envFromSource = append(envFromSource,
				corev1.EnvFromSource{
					ConfigMapRef: &corev1.ConfigMapEnvSource{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: env.ConfigMap,
						},
					},
				},
			)
		} else if len(env.Secret) > 0 {
			envFromSource = append(envFromSource,
				corev1.EnvFromSource{
					SecretRef: &corev1.SecretEnvSource{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: env.Secret,
						},
					},
				},
			)
		}
	}

	return envFromSource
}

func getContainerPorts(application *skiperatorv1alpha1.Application) []corev1.ContainerPort {

	containerPorts := []corev1.ContainerPort{
		{
			Name:          "main",
			ContainerPort: int32(application.Spec.Port),
		},
	}

	for _, port := range application.Spec.AdditionalPorts {
		containerPorts = append(containerPorts, corev1.ContainerPort{
			ContainerPort: port.Port,
			Name:          port.Name,
			Protocol:      port.Protocol,
		})
	}

	return containerPorts
}
