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

		deployment.Spec.Template.ObjectMeta.Annotations = map[string]string{
			"argocd.argoproj.io/sync-options": "Prune=false",
			"prometheus.io/scrape":            "true",
		}

		// add an external link to argocd
		ingresses := application.Spec.Ingresses
		if len(ingresses) > 0 {
			deployment.ObjectMeta.Annotations[AnnotationKeyLinkPrefix] = fmt.Sprintf("https://%s", ingresses[0])
		}

		labels := map[string]string{"app": application.Name}
		deployment.Spec.Template.ObjectMeta.Labels = labels
		deployment.Spec.Selector = &metav1.LabelSelector{MatchLabels: labels}

		var replicas = int32(application.Spec.Replicas.Min)
		if replicas == 0 {
			replicas = 1
		}
		deployment.Spec.Replicas = &replicas
		revisionHistoryLimit := int32(2)
		deployment.Spec.RevisionHistoryLimit = &revisionHistoryLimit

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
		container.Ports[0].Name = "main"

		for _, port := range application.Spec.AdditionalPorts {
			container.Ports = append(container.Ports, corev1.ContainerPort{
				ContainerPort: port.Port,
				Name:          port.Name,
				Protocol:      port.Protocol,
			})
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
			}
		}

		container.Resources = corev1.ResourceRequirements{
			Limits:   application.Spec.Resources.Limits,
			Requests: application.Spec.Resources.Requests,
		}

		numberOfVolumes := len(application.Spec.FilesFrom) + 1
		if application.Spec.GCP != nil {
			numberOfVolumes = numberOfVolumes + 1
		}

		deployment.Spec.Template.Spec.Volumes = make([]corev1.Volume, numberOfVolumes)
		container.VolumeMounts = make([]corev1.VolumeMount, numberOfVolumes)
		volumes := deployment.Spec.Template.Spec.Volumes
		volumeMounts := container.VolumeMounts

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
			container.Env = append(application.Spec.Env, envVar)

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
			container.Env = application.Spec.Env
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

		return nil
	})

	r.SetControllerFinishedOutcome(ctx, application, controllerName, err)

	return reconcile.Result{}, err
}
