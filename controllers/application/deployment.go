package applicationcontroller

import (
	"context"
	"fmt"

	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/util"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// Adding an argocd external link constant
const (
	AnnotationKeyLinkPrefix = "link.argocd.argoproj.io/external-link"
)

func (r *ApplicationReconciler) reconcileDeployment(ctx context.Context, application *skiperatorv1alpha1.Application) (reconcile.Result, error) {
	controllerName := "Deployment"
	r.SetControllerProgressing(ctx, application, controllerName)

	deployment := appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Namespace: application.Namespace, Name: application.Name}}

	_, err := ctrlutil.CreateOrPatch(ctx, r.GetClient(), &deployment, func() error {
		// Set application as owner of the deployment
		err := ctrlutil.SetControllerReference(application, &deployment, r.GetScheme())
		if err != nil {
			r.SetControllerError(ctx, application, controllerName, err)
			return err
		}

		r.SetLabelsFromApplication(ctx, &deployment, *application)
		util.SetCommonAnnotations(&deployment)

		labels := map[string]string{"app": application.Name}
		deployment.Spec.Template.ObjectMeta.Labels = labels

		deployment.Spec.Template.ObjectMeta.Annotations = map[string]string{
			"argocd.argoproj.io/sync-options": "Prune=false",
			"prometheus.io/scrape":            "true",
		}

		// add an external link to argocd
		ingresses := application.Spec.Ingresses
		if len(ingresses) > 0 {
			deployment.ObjectMeta.Annotations[AnnotationKeyLinkPrefix] = fmt.Sprintf("https://%s", ingresses[0])
		}

		skiperatorContainer := corev1.Container{
			Name:            application.Name,
			Image:           application.Spec.Image,
			ImagePullPolicy: corev1.PullAlways,
			Command:         application.Spec.Command,
			SecurityContext: &corev1.SecurityContext{
				SeccompProfile: &corev1.SeccompProfile{
					Type: corev1.SeccompProfileTypeRuntimeDefault,
				},
				Privileged:               util.PointTo(false),
				AllowPrivilegeEscalation: util.PointTo(false),
				ReadOnlyRootFilesystem:   util.PointTo(true),
				RunAsUser:                &util.SkiperatorUser,
				RunAsGroup:               &util.SkiperatorUser,
			},
			Ports:   getContainerPorts(application),
			EnvFrom: getEnvFrom(application.Spec.EnvFrom),
			Resources: corev1.ResourceRequirements{
				Limits:   application.Spec.Resources.Limits,
				Requests: application.Spec.Resources.Requests,
			},
		}

		numberOfVolumes := len(application.Spec.FilesFrom) + 1
		if application.Spec.GCP != nil {
			numberOfVolumes = numberOfVolumes + 1
		}

		deployment.Spec.Template.Spec.Volumes = make([]corev1.Volume, numberOfVolumes)
		skiperatorContainer.VolumeMounts = make([]corev1.VolumeMount, numberOfVolumes)
		volumes := deployment.Spec.Template.Spec.Volumes
		volumeMounts := skiperatorContainer.VolumeMounts

		volumes[0] = corev1.Volume{Name: "tmp", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}}
		volumeMounts[0] = corev1.VolumeMount{Name: "tmp", MountPath: "/tmp"}

		for i, file := range application.Spec.FilesFrom {
			if len(file.ConfigMap) > 0 {
				volumes[i+1].Name = file.ConfigMap
				volumes[i+1].ConfigMap = &corev1.ConfigMapVolumeSource{}
				volumes[i+1].ConfigMap.Name = file.ConfigMap
			} else if len(file.Secret) > 0 {
				volumes[i+1].Name = file.Secret
				volumes[i+1].Secret = &corev1.SecretVolumeSource{}
				volumes[i+1].Secret.SecretName = file.Secret
			} else if len(file.EmptyDir) > 0 {
				volumes[i+1].Name = file.EmptyDir
				volumes[i+1].EmptyDir = &corev1.EmptyDirVolumeSource{}
			} else if len(file.PersistentVolumeClaim) > 0 {
				volumes[i+1].Name = file.PersistentVolumeClaim
				volumes[i+1].PersistentVolumeClaim = &corev1.PersistentVolumeClaimVolumeSource{}
				volumes[i+1].PersistentVolumeClaim.ClaimName = file.PersistentVolumeClaim
			}

			volumeMounts[i+1] = corev1.VolumeMount{Name: volumes[i+1].Name, MountPath: file.MountPath}
		}

		if application.Spec.GCP != nil {
			gcpIdentityConfigMapNamespacedName := types.NamespacedName{Namespace: "skiperator-system", Name: "gcp-identity-config"}
			gcpIdentityConfigMap := corev1.ConfigMap{}

			gcpIdentityConfigMap, err := util.GetConfigMap(r.GetClient(), ctx, gcpIdentityConfigMapNamespacedName)

			if !util.ErrIsMissingOrNil(
				r.GetRecorder(),
				err,
				"Cannot find configmap named "+gcpIdentityConfigMapNamespacedName.Name+" in namespace "+gcpIdentityConfigMapNamespacedName.Namespace,
				application,
			) {
				r.SetControllerError(ctx, application, controllerName, err)
				return err
			}

			envVar := corev1.EnvVar{
				Name:  "GOOGLE_APPLICATION_CREDENTIALS",
				Value: "/var/run/secrets/tokens/gcp-ksa/google-application-credentials.json",
			}
			skiperatorContainer.Env = append(application.Spec.Env, envVar)

			volumeMounts[numberOfVolumes-1].Name = "gcp-ksa"
			volumeMounts[numberOfVolumes-1].MountPath = "/var/run/secrets/tokens/gcp-ksa"
			volumeMounts[numberOfVolumes-1].ReadOnly = true

			twoDaysSec := int64(172800)
			optionalBool := false
			defaultModeValue := int32(420)

			volumes[numberOfVolumes-1].Name = "gcp-ksa"
			volumes[numberOfVolumes-1].Projected = &corev1.ProjectedVolumeSource{}
			volumes[numberOfVolumes-1].Projected.DefaultMode = &defaultModeValue
			volumes[numberOfVolumes-1].Projected.Sources = make([]corev1.VolumeProjection, 2)
			volumes[numberOfVolumes-1].Projected.Sources[0].ServiceAccountToken = &corev1.ServiceAccountTokenProjection{
				Path:              "token",
				Audience:          gcpIdentityConfigMap.Data["workloadIdentityPool"],
				ExpirationSeconds: &twoDaysSec,
			}
			volumes[numberOfVolumes-1].Projected.Sources[1].ConfigMap = &corev1.ConfigMapProjection{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: application.Name + "-gcp-auth",
				},
				Optional: &optionalBool,
				Items:    make([]corev1.KeyToPath, 1),
			}
			volumes[numberOfVolumes-1].Projected.Sources[1].ConfigMap.Items[0].Key = "config"
			volumes[numberOfVolumes-1].Projected.Sources[1].ConfigMap.Items[0].Path = "google-application-credentials.json"
		} else {
			skiperatorContainer.Env = application.Spec.Env
		}

		if application.Spec.Readiness != nil {
			skiperatorContainer.ReadinessProbe = &corev1.Probe{}
			skiperatorContainer.ReadinessProbe.InitialDelaySeconds = int32(application.Spec.Readiness.InitialDelay)
			skiperatorContainer.ReadinessProbe.TimeoutSeconds = int32(application.Spec.Readiness.Timeout)
			skiperatorContainer.ReadinessProbe.FailureThreshold = int32(application.Spec.Readiness.FailureThreshold)

			skiperatorContainer.ReadinessProbe.HTTPGet = &corev1.HTTPGetAction{}
			skiperatorContainer.ReadinessProbe.HTTPGet.Port = intstr.FromInt(int(application.Spec.Readiness.Port))
			skiperatorContainer.ReadinessProbe.HTTPGet.Path = application.Spec.Readiness.Path
		}
		if application.Spec.Liveness != nil {
			skiperatorContainer.LivenessProbe = &corev1.Probe{}
			skiperatorContainer.LivenessProbe.InitialDelaySeconds = int32(application.Spec.Liveness.InitialDelay)
			skiperatorContainer.LivenessProbe.TimeoutSeconds = int32(application.Spec.Liveness.Timeout)
			skiperatorContainer.LivenessProbe.FailureThreshold = int32(application.Spec.Liveness.FailureThreshold)

			skiperatorContainer.LivenessProbe.HTTPGet = &corev1.HTTPGetAction{}
			skiperatorContainer.LivenessProbe.HTTPGet.Port = intstr.FromInt(int(application.Spec.Liveness.Port))
			skiperatorContainer.LivenessProbe.HTTPGet.Path = application.Spec.Liveness.Path
		}
		if application.Spec.Startup != nil {
			skiperatorContainer.StartupProbe = &corev1.Probe{}
			skiperatorContainer.StartupProbe.InitialDelaySeconds = int32(application.Spec.Startup.InitialDelay)
			skiperatorContainer.StartupProbe.TimeoutSeconds = int32(application.Spec.Startup.Timeout)
			skiperatorContainer.StartupProbe.FailureThreshold = int32(application.Spec.Startup.FailureThreshold)

			skiperatorContainer.StartupProbe.HTTPGet = &corev1.HTTPGetAction{}
			skiperatorContainer.StartupProbe.HTTPGet.Port = intstr.FromInt(int(application.Spec.Startup.Port))
			skiperatorContainer.StartupProbe.HTTPGet.Path = application.Spec.Startup.Path
		}

		deployment.Spec = appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{MatchLabels: labels},
			Replicas: getReplicasFromAppSpec(application.Spec.Replicas.Min),
			Strategy: appsv1.DeploymentStrategy{
				Type:          appsv1.DeploymentStrategyType(application.Spec.Strategy.Type),
				RollingUpdate: getRollingUpdateStrategy(application.Spec.Strategy.Type),
			},
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						skiperatorContainer,
					},

					// TODO: Make this as part of operator in a safe way
					ImagePullSecrets: []corev1.LocalObjectReference{{Name: "github-auth"}},
					SecurityContext: &corev1.PodSecurityContext{
						SupplementalGroups: []int64{util.SkiperatorUser},
						FSGroup:            &util.SkiperatorUser,
					},
					ServiceAccountName: application.Name,
				},
			},
		}

		return nil
	})

	r.SetControllerFinishedOutcome(ctx, application, controllerName, err)

	return reconcile.Result{}, err
}

func getEnvFrom(envFromApplication []skiperatorv1alpha1.EnvFrom) []corev1.EnvFromSource {
	envFromSource := []corev1.EnvFromSource{}

	for i, env := range envFromApplication {
		if len(env.ConfigMap) > 0 {
			envFromSource[i] = corev1.EnvFromSource{
				ConfigMapRef: &corev1.ConfigMapEnvSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: env.ConfigMap,
					},
				},
			}
		} else if len(env.Secret) > 0 {
			envFromSource[i] = corev1.EnvFromSource{
				SecretRef: &corev1.SecretEnvSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: env.Secret,
					},
				},
			}
		}
	}

	return envFromSource
}

func getContainerPorts(application *skiperatorv1alpha1.Application) []corev1.ContainerPort {

	containerPorts := []corev1.ContainerPort{
		corev1.ContainerPort{
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

func getRollingUpdateStrategy(updateStrategy string) *appsv1.RollingUpdateDeployment {
	if updateStrategy == "Recreate" {
		return nil
	}

	return &appsv1.RollingUpdateDeployment{}
}

func getReplicasFromAppSpec(appReplicas uint) *int32 {
	var replicas = int32(appReplicas)
	if replicas == 0 {
		minReplicas := int32(1)
		return &minReplicas
	}

	return &replicas
}
