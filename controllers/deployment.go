package controllers

import (
	"context"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

//+kubebuilder:rbac:groups=skiperator.kartverket.no,resources=applications,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete

type DeploymentReconciler struct {
	client client.Client
	scheme *runtime.Scheme
}

func (r *DeploymentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.client = mgr.GetClient()
	r.scheme = mgr.GetScheme()

	return ctrl.NewControllerManagedBy(mgr).
		For(&skiperatorv1alpha1.Application{}).
		Owns(&appsv1.Deployment{}).
		Complete(r)
}

func (r *DeploymentReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	// Fetch application and fill defaults
	application := skiperatorv1alpha1.Application{}
	err := r.client.Get(ctx, req.NamespacedName, &application)
	if err != nil {
		err = client.IgnoreNotFound(err)
		return reconcile.Result{}, err
	}
	application.FillDefaults()

	deployment := appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Namespace: req.Namespace, Name: req.Name}}
	_, err = ctrlutil.CreateOrPatch(ctx, r.client, &deployment, func() error {
		// Set application as owner of the deployment
		err = ctrlutil.SetControllerReference(&application, &deployment, r.scheme)
		if err != nil {
			return err
		}

		deployment.Spec.Template.ObjectMeta.Annotations = map[string]string{"prometheus.io/scrape": "true"}

		labels := map[string]string{"app": application.Name}
		deployment.Spec.Template.ObjectMeta.Labels = labels
		deployment.Spec.Selector = &metav1.LabelSelector{MatchLabels: labels}

		var replicas = int32(application.Spec.Replicas.Min)
		if replicas == 0 {
			replicas = 1
		}
		deployment.Spec.Replicas = &replicas

		deployment.Spec.Template.Spec.Containers = make([]corev1.Container, 1)
		container := &deployment.Spec.Template.Spec.Containers[0]
		container.Name = application.Name

		// TODO: Make this as part of operator in a safe way
		deployment.Spec.Template.Spec.ImagePullSecrets = []corev1.LocalObjectReference{{Name: "github-auth"}}
		container.Image = application.Spec.Image
		container.ImagePullPolicy = corev1.PullAlways
		container.Command = application.Spec.Command

		var uid int64 = 150 // TODO: 65534? Evnt. hashed? Random?
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

		container.Env = application.Spec.Env
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

		container.Resources = application.Spec.Resources

		deployment.Spec.Template.Spec.Volumes = make([]corev1.Volume, len(application.Spec.FilesFrom)+1)
		container.VolumeMounts = make([]corev1.VolumeMount, len(application.Spec.FilesFrom)+1)
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
				volumes[i+1].Name = file.PersistentVolumeClaim
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

		return nil
	})
	return reconcile.Result{}, err
}
