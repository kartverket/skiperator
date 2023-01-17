package applicationcontroller

import (
	"context"
	"fmt"
	"strings"

	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	util "github.com/kartverket/skiperator/pkg/util"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *ApplicationReconciler) reconcileDeployment(ctx context.Context, application *skiperatorv1alpha1.Application) (reconcile.Result, error) {
	controllerName := "Deployment"
	r.SetControllerProgressing(ctx, application, controllerName)

	gcpIdentityConfigMap := corev1.ConfigMap{}

	if application.Spec.GCP != nil {
		err := r.GetClient().Get(ctx, types.NamespacedName{Namespace: "skiperator-system", Name: "gcp-identity-config"}, &gcpIdentityConfigMap)
		if errors.IsNotFound(err) {
			r.GetRecorder().Eventf(
				application,
				corev1.EventTypeWarning, "Missing",
				"Cannot find configmap named gcp-identity-config in namespace skiperator-system",
			)
		} else if err != nil {
			r.SetControllerError(ctx, application, controllerName, err)
			return reconcile.Result{}, err
		}
	}

	deployment := appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Namespace: application.Namespace, Name: application.Name}}
	_, err := ctrlutil.CreateOrPatch(ctx, r.GetClient(), &deployment, func() error {
		// Set application as owner of the deployment
		err := ctrlutil.SetControllerReference(application, &deployment, r.GetScheme())
		if err != nil {
			r.SetControllerError(ctx, application, controllerName, err)
			return err
		}

		r.SetLabelsFromApplication(ctx, &deployment, *application)

		deployment.Spec.Template.ObjectMeta.Annotations = map[string]string{"prometheus.io/scrape": "true"}

		labels := map[string]string{"app": application.Name}
		deployment.Spec.Template.ObjectMeta.Labels = labels
		deployment.Spec.Selector = &metav1.LabelSelector{MatchLabels: labels}

		var replicas = int32(application.Spec.Replicas.Min)
		if replicas == 0 {
			replicas = 1
		}
		deployment.Spec.Replicas = &replicas

		deployment.Spec.Strategy.Type = appsv1.DeploymentStrategyType(application.Spec.Strategy.Type)
		if application.Spec.Strategy.Type == "Recreate" {
			deployment.Spec.Strategy.RollingUpdate = nil
		}

		deployment.Spec.Template.Spec.Containers = make([]corev1.Container, 1)
		container := &deployment.Spec.Template.Spec.Containers[0]
		container.Name = application.Name

		// TODO: Make this as part of operator in a safe way
		deployment.Spec.Template.Spec.ImagePullSecrets = []corev1.LocalObjectReference{{Name: "github-auth"}}
		container.Image = application.Spec.Image
		container.ImagePullPolicy = corev1.PullAlways
		container.Command = application.Spec.Command

		var uid int64 = 150 // TODO: 65534? Evnt. hashed? Random?
		deployment.Spec.Template.Spec.ServiceAccountName = application.Name
		deployment.Spec.Template.Spec.SecurityContext = &corev1.PodSecurityContext{}
		deployment.Spec.Template.Spec.SecurityContext.SupplementalGroups = []int64{uid}
		deployment.Spec.Template.Spec.SecurityContext.FSGroup = &uid

		yes := true
		no := false
		container.SecurityContext = &corev1.SecurityContext{}
		container.SecurityContext.SeccompProfile = &corev1.SeccompProfile{}
		container.SecurityContext.SeccompProfile.Type = "RuntimeDefault"
		container.SecurityContext.Privileged = &no
		container.SecurityContext.AllowPrivilegeEscalation = &no
		container.SecurityContext.ReadOnlyRootFilesystem = &yes
		container.SecurityContext.RunAsUser = &uid
		container.SecurityContext.RunAsGroup = &uid

		container.Ports = make([]corev1.ContainerPort, 1)
		container.Ports[0].ContainerPort = int32(application.Spec.Port)

		//Adding env for GCP authentication
		if application.Spec.GCP != nil {
			envVar := corev1.EnvVar{
				Name:  "GOOGLE_APPLICATION_CREDENTIALS",
				Value: "/var/run/secrets/tokens/gcp-ksa/google-application-credentials.json",
			}
			container.Env = append(application.Spec.Env, envVar)
		} else {
			container.Env = application.Spec.Env
		}

		container.EnvFrom = make([]corev1.EnvFromSource, len(application.Spec.EnvFrom))
		envFrom := container.EnvFrom

		for i, env := range application.Spec.EnvFrom {
			if len(env.ConfigMap) > 0 {
				envFrom[i].ConfigMapRef = &corev1.ConfigMapEnvSource{}
				envFrom[i].ConfigMapRef.LocalObjectReference.Name = env.ConfigMap
			} else if len(env.Secret) > 0 {
				envFrom[i].SecretRef = &corev1.SecretEnvSource{}
				envFrom[i].SecretRef.LocalObjectReference.Name = env.Secret
			} else if len(env.GcpSecretManager) > 0 {
				secretName := strings.Split(env.GcpSecretManager, "/")[3]
				envFrom[i].SecretRef = &corev1.SecretEnvSource{}
				envFrom[i].SecretRef.LocalObjectReference.Name = fmt.Sprintf("gcp-sm-%s", secretName)
			}
		}

		container.Resources = corev1.ResourceRequirements{
			Limits:   application.Spec.Resources.Limits,
			Requests: application.Spec.Resources.Requests,
		}

		csiVolumes := getCsiVolumes(application)

		volumes := make([]corev1.Volume, 1)
		volumeMounts := make([]corev1.VolumeMount, 1)

		volumes[0] = corev1.Volume{Name: "tmp", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}}
		volumeMounts[0] = corev1.VolumeMount{Name: "tmp", MountPath: "/tmp"}

		for _, file := range application.Spec.FilesFrom {
			volume := corev1.Volume{}
			if len(file.ConfigMap) > 0 {
				volume.Name = file.ConfigMap
				volume.ConfigMap = &corev1.ConfigMapVolumeSource{}
				volume.ConfigMap.Name = file.ConfigMap
			} else if len(file.Secret) > 0 {
				volume.Name = file.Secret
				volume.Secret = &corev1.SecretVolumeSource{}
				volume.Secret.SecretName = file.Secret
			} else if len(file.EmptyDir) > 0 {
				volume.Name = file.EmptyDir
				volume.EmptyDir = &corev1.EmptyDirVolumeSource{}
			} else if len(file.PersistentVolumeClaim) > 0 {
				volume.Name = file.PersistentVolumeClaim
				volume.PersistentVolumeClaim = &corev1.PersistentVolumeClaimVolumeSource{}
				volume.PersistentVolumeClaim.ClaimName = file.PersistentVolumeClaim
			}

			if len(file.GcpSecretManager) == 0 {
				volumes = append(volumes, volume)
				volumeMount := corev1.VolumeMount{Name: volume.Name, MountPath: file.MountPath}
				volumeMounts = append(volumeMounts, volumeMount)
			} else {
				hash := util.GenerateHashFromName(file.GcpSecretManager)
				name := fmt.Sprintf("%x", hash)
				name = fmt.Sprintf("gcp-sm-%s", name)
				volumeMount := corev1.VolumeMount{Name: name, MountPath: file.MountPath}
				volumeMounts = append(volumeMounts, volumeMount)
			}
		}

		// EnvFrom Secret Manager requires a volume to trigger CSI plugin so this must be separate
		// from the logic above which only handles FilesFrom
		for _, secretManagerReference := range csiVolumes {
			volume := corev1.Volume{}
			hash := util.GenerateHashFromName(secretManagerReference)
			name := fmt.Sprintf("%x", hash)

			volume.Name = fmt.Sprintf("gcp-sm-%s", name)
			volume.CSI = &corev1.CSIVolumeSource{}
			volume.CSI.Driver = "secrets-store.csi.k8s.io"
			volume.CSI.ReadOnly = &yes
			volume.CSI.VolumeAttributes = make(map[string]string)
			volume.CSI.VolumeAttributes["secretProviderClass"] = fmt.Sprintf("%s-%s", application.Name, name)

			volumes = append(volumes, volume)
		}

		if application.Spec.GCP != nil {
			volume := corev1.Volume{}
			volumeMount := corev1.VolumeMount{}

			volumeMount.Name = "gcp-ksa"
			volumeMount.MountPath = "/var/run/secrets/tokens/gcp-ksa"
			volumeMount.ReadOnly = true

			twoDaysSec := int64(172800)
			optionalBool := false
			defaultModeValue := int32(420)

			volume.Name = "gcp-ksa"
			volume.Projected = &corev1.ProjectedVolumeSource{}
			volume.Projected.DefaultMode = &defaultModeValue
			volume.Projected.Sources = make([]corev1.VolumeProjection, 2)
			volume.Projected.Sources[0].ServiceAccountToken = &corev1.ServiceAccountTokenProjection{
				Path:              "token",
				Audience:          gcpIdentityConfigMap.Data["workloadIdentityPool"],
				ExpirationSeconds: &twoDaysSec,
			}
			volume.Projected.Sources[1].ConfigMap = &corev1.ConfigMapProjection{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: application.Name + "-gcp-auth",
				},
				Optional: &optionalBool,
				Items:    make([]corev1.KeyToPath, 1),
			}
			volume.Projected.Sources[1].ConfigMap.Items[0].Key = "config"
			volume.Projected.Sources[1].ConfigMap.Items[0].Path = "google-application-credentials.json"

			volumes = append(volumes, volume)
			volumeMounts = append(volumeMounts, volumeMount)
		}

		if application.Spec.Readiness != nil {
			container.ReadinessProbe = &corev1.Probe{}
			container.ReadinessProbe.InitialDelaySeconds = int32(application.Spec.Readiness.InitialDelay)
			container.ReadinessProbe.TimeoutSeconds = int32(application.Spec.Readiness.Timeout)
			container.ReadinessProbe.FailureThreshold = int32(application.Spec.Readiness.FailureThreshold)

			container.ReadinessProbe.HTTPGet = &corev1.HTTPGetAction{}
			container.ReadinessProbe.HTTPGet.Port = intstr.FromInt(int(application.Spec.Readiness.Port))
			container.ReadinessProbe.HTTPGet.Path = application.Spec.Readiness.Path
		}
		if application.Spec.Liveness != nil {
			container.LivenessProbe = &corev1.Probe{}
			container.LivenessProbe.InitialDelaySeconds = int32(application.Spec.Liveness.InitialDelay)
			container.LivenessProbe.TimeoutSeconds = int32(application.Spec.Liveness.Timeout)
			container.LivenessProbe.FailureThreshold = int32(application.Spec.Liveness.FailureThreshold)

			container.LivenessProbe.HTTPGet = &corev1.HTTPGetAction{}
			container.LivenessProbe.HTTPGet.Port = intstr.FromInt(int(application.Spec.Liveness.Port))
			container.LivenessProbe.HTTPGet.Path = application.Spec.Liveness.Path
		}
		if application.Spec.Startup != nil {
			container.StartupProbe = &corev1.Probe{}
			container.StartupProbe.InitialDelaySeconds = int32(application.Spec.Startup.InitialDelay)
			container.StartupProbe.TimeoutSeconds = int32(application.Spec.Startup.Timeout)
			container.StartupProbe.FailureThreshold = int32(application.Spec.Startup.FailureThreshold)

			container.StartupProbe.HTTPGet = &corev1.HTTPGetAction{}
			container.StartupProbe.HTTPGet.Port = intstr.FromInt(int(application.Spec.Startup.Port))
			container.StartupProbe.HTTPGet.Path = application.Spec.Startup.Path
		}

		deployment.Spec.Template.Spec.Volumes = volumes
		container.VolumeMounts = volumeMounts

		return nil
	})

	r.SetControllerFinishedOutcome(ctx, application, controllerName, err)

	return reconcile.Result{}, err
}

func getCsiVolumes(application *skiperatorv1alpha1.Application) []string {
	result := []string{}

	for _, envFrom := range application.Spec.EnvFrom {
		value := envFrom.GcpSecretManager
		if len(value) > 0 && !util.Contains(result, value) {
			result = append(result, envFrom.GcpSecretManager)
		}
	}
	for _, fileFrom := range application.Spec.FilesFrom {
		value := fileFrom.GcpSecretManager
		if len(value) > 0 && !util.Contains(result, value) {
			result = append(result, fileFrom.GcpSecretManager)
		}
	}

	return result
}
