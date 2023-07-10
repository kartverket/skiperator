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
	AnnotationKeyLinkPrefix                = "link.argocd.argoproj.io/external-link"
	DefaultDigdiratorMaskinportenMountPath = "/var/run/secrets/skip/maskinporten"
	DefaultDigdiratorIDportenMountPath     = "/var/run/secrets/skip/idporten"
)

func (r *ApplicationReconciler) reconcileDeployment(ctx context.Context, application *skiperatorv1alpha1.Application) (reconcile.Result, error) {
	controllerName := "Deployment"
	r.SetControllerProgressing(ctx, application, controllerName)

	deployment := appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: application.Namespace,
			Name:      application.Name,
		},
	}

	_, err := ctrlutil.CreateOrPatch(ctx, r.GetClient(), &deployment, func() error {
		// Set application as owner of the deployment
		err := ctrlutil.SetControllerReference(application, &deployment, r.GetScheme())
		if err != nil {
			r.SetControllerError(ctx, application, controllerName, err)
			return err
		}

		r.SetLabelsFromApplication(ctx, &deployment, *application)
		util.SetCommonAnnotations(&deployment)

		skiperatorContainer := corev1.Container{
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
			Ports:   getContainerPorts(application),
			EnvFrom: getEnvFrom(application.Spec.EnvFrom),
			Resources: corev1.ResourceRequirements{
				Limits:   application.Spec.Resources.Limits,
				Requests: application.Spec.Resources.Requests,
			},
			Env:            application.Spec.Env,
			ReadinessProbe: getProbe(application.Spec.Readiness),
			LivenessProbe:  getProbe(application.Spec.Liveness),
			StartupProbe:   getProbe(application.Spec.Startup),
		}

		podVolumes, containerVolumeMounts := getContainerVolumeMountsAndPodVolumes(application)
		podVolumes, containerVolumeMounts, err = r.appendGCPVolumeMount(application, ctx, &skiperatorContainer, containerVolumeMounts, podVolumes)

		if err != nil {
			r.SetControllerError(ctx, application, controllerName, err)
			return err
		}

		podVolumes, containerVolumeMounts, err = r.appendMaskinportenSecretVolumeMount(application, ctx, &skiperatorContainer, containerVolumeMounts, podVolumes)
		if err != nil {
			r.SetControllerError(ctx, application, controllerName, err)
			return err
		}

		podVolumes, containerVolumeMounts, err = r.appendIDportenSecretVolumeMount(application, ctx, &skiperatorContainer, containerVolumeMounts, podVolumes)
		if err != nil {
			r.SetControllerError(ctx, application, controllerName, err)
			return err
		}

		skiperatorContainer.VolumeMounts = containerVolumeMounts

		labels := util.GetApplicationSelector(application.Name)

		deployment.Spec = appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{MatchLabels: labels},
			Replicas: getReplicasFromAppSpec(application.Spec.Replicas.Min),
			Strategy: appsv1.DeploymentStrategy{
				Type:          appsv1.DeploymentStrategyType(application.Spec.Strategy.Type),
				RollingUpdate: getRollingUpdateStrategy(application.Spec.Strategy.Type),
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
					Annotations: map[string]string{
						"argocd.argoproj.io/sync-options": "Prune=false",
						"prometheus.io/scrape":            "true",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						skiperatorContainer,
					},

					// TODO: Make this as part of operator in a safe way
					ImagePullSecrets: []corev1.LocalObjectReference{{Name: "github-auth"}},
					SecurityContext: &corev1.PodSecurityContext{
						SupplementalGroups: []int64{util.SkiperatorUser},
						FSGroup:            util.PointTo(util.SkiperatorUser),
						SeccompProfile: &corev1.SeccompProfile{
							Type: corev1.SeccompProfileTypeRuntimeDefault,
						},
					},
					ServiceAccountName: application.Name,
					Volumes:            podVolumes,
					PriorityClassName:  fmt.Sprintf("skip-%s", application.Spec.Priority),
				},
			},
			RevisionHistoryLimit: util.PointTo(int32(2)),
		}

		// add an external link to argocd
		ingresses := application.Spec.Ingresses
		if len(ingresses) > 0 {
			deployment.ObjectMeta.Annotations[AnnotationKeyLinkPrefix] = fmt.Sprintf("https://%s", ingresses[0])
		}

		return nil
	})

	r.SetControllerFinishedOutcome(ctx, application, controllerName, err)

	return reconcile.Result{}, err
}

func getProbe(appProbe *skiperatorv1alpha1.Probe) *corev1.Probe {
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
func (r ApplicationReconciler) appendMaskinportenSecretVolumeMount(application *skiperatorv1alpha1.Application, ctx context.Context, skiperatorContainer *corev1.Container, volumeMounts []corev1.VolumeMount, volumes []corev1.Volume) ([]corev1.Volume, []corev1.VolumeMount, error) {
	if maskinportenSpecifiedInSpec(application.Spec.Maskinporten) {

		// TODO: check if client is created (?)

		secretName, err := util.GetSecretName("maskinporten", application.Name)

		if err != nil {
			r.SetControllerError(ctx, application, controllerName, err)
			return volumes, volumeMounts, err
		}

		// TODO: Check if secret exits

		volumeMounts = append(volumeMounts, fromFilesVolumeMount(secretName, DefaultDigdiratorMaskinportenMountPath, true))
		volumes = append(volumes, fromFilesVolume(secretName, secretName))
	}
	return volumes, volumeMounts, nil
}

func (r ApplicationReconciler) appendIDportenSecretVolumeMount(application *skiperatorv1alpha1.Application, ctx context.Context, skiperatorContainer *corev1.Container, volumeMounts []corev1.VolumeMount, volumes []corev1.Volume) ([]corev1.Volume, []corev1.VolumeMount, error) {
	if idportenSpecifiedInSpec(application.Spec.IDPorten) {
		// TODO: check if client is created (?)

		secretName, err := util.GetSecretName("idporten", application.Name)

		if err != nil {
			r.SetControllerError(ctx, application, controllerName, err)
			return volumes, volumeMounts, err
		}

		// TODO: Check if secret exits

		volumeMounts = append(volumeMounts, fromFilesVolumeMount(secretName, DefaultDigdiratorIDportenMountPath, true))
		volumes = append(volumes, fromFilesVolume(secretName, secretName))
	}
	return volumes, volumeMounts, nil
}

func (r ApplicationReconciler) appendGCPVolumeMount(application *skiperatorv1alpha1.Application, ctx context.Context, skiperatorContainer *corev1.Container, volumeMounts []corev1.VolumeMount, volumes []corev1.Volume) ([]corev1.Volume, []corev1.VolumeMount, error) {
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
			return volumes, volumeMounts, err
		}

		envVar := corev1.EnvVar{
			Name:  "GOOGLE_APPLICATION_CREDENTIALS",
			Value: "/var/run/secrets/tokens/gcp-ksa/google-application-credentials.json",
		}
		skiperatorContainer.Env = append(application.Spec.Env, envVar)

		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      "gcp-ksa",
			MountPath: "/var/run/secrets/tokens/gcp-ksa",
			ReadOnly:  true,
		})

		twoDaysInSeconds := int64(172800)

		volumes = append(volumes, corev1.Volume{
			Name: "gcp-ksa",
			VolumeSource: corev1.VolumeSource{
				Projected: &corev1.ProjectedVolumeSource{
					DefaultMode: util.PointTo(int32(420)),
					Sources: []corev1.VolumeProjection{
						{
							ServiceAccountToken: &corev1.ServiceAccountTokenProjection{
								Path:              "token",
								Audience:          gcpIdentityConfigMap.Data["workloadIdentityPool"],
								ExpirationSeconds: &twoDaysInSeconds,
							},
						},
						{
							ConfigMap: &corev1.ConfigMapProjection{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: application.Name + "-gcp-auth",
								},
								Optional: util.PointTo(false),
								Items: []corev1.KeyToPath{
									{
										Key:  "config",
										Path: "google-application-credentials.json",
									},
								},
							},
						},
					},
				},
			},
		})
	}

	return volumes, volumeMounts, nil
}

func getContainerVolumeMountsAndPodVolumes(application *skiperatorv1alpha1.Application) ([]corev1.Volume, []corev1.VolumeMount) {
	containerVolumeMounts := []corev1.VolumeMount{
		{
			Name:      "tmp",
			MountPath: "/tmp",
		},
	}

	podVolumes := []corev1.Volume{
		{
			Name: "tmp",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		},
	}

	for _, file := range application.Spec.FilesFrom {
		volume := corev1.Volume{}
		if len(file.ConfigMap) > 0 {
			volume = corev1.Volume{
				Name: file.ConfigMap,
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: file.ConfigMap,
						},
					},
				},
			}
		} else if len(file.Secret) > 0 {
			volume = corev1.Volume{
				Name: file.Secret,
				VolumeSource: corev1.VolumeSource{
					Secret: &corev1.SecretVolumeSource{
						SecretName: file.Secret,
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

		podVolumes = append(podVolumes, volume)
		containerVolumeMounts = append(containerVolumeMounts, corev1.VolumeMount{
			Name:      volume.Name,
			MountPath: file.MountPath,
		})
	}

	return podVolumes, containerVolumeMounts
}

func fromFilesVolumeMount(secretName string, mountPath string, readOnly bool) corev1.VolumeMount {
	return corev1.VolumeMount{
		Name:      secretName,
		MountPath: mountPath,
		ReadOnly:  readOnly,
	}
}

func fromFilesVolume(volumeName string, secretName string) corev1.Volume {
	return corev1.Volume{
		Name: volumeName,
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: secretName,
				Items:      nil,
			},
		},
	}
}

func getEnvFrom(envFromApplication []skiperatorv1alpha1.EnvFrom) []corev1.EnvFromSource {
	envFromSource := []corev1.EnvFromSource{}

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
