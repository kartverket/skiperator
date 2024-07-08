package deployment

import (
	"context"
	goerrors "errors"
	"fmt"
	"github.com/kartverket/skiperator/pkg/log"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/idporten"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/maskinporten"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/pod"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/resourceutils"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/volume"
	"k8s.io/client-go/rest"
	"strings"

	"github.com/go-logr/logr"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/gcp"
	"github.com/kartverket/skiperator/pkg/util"
	"golang.org/x/exp/maps"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	AnnotationKeyLinkPrefix                = "link.argocd.argoproj.io/external-link"
	DefaultDigdiratorMaskinportenMountPath = "/var/run/secrets/skip/maskinporten"
	DefaultDigdiratorIDportenMountPath     = "/var/run/secrets/skip/idporten"
)

func Generate(ctx context.Context, application *skiperatorv1alpha1.Application, istioEnabled bool, workloadIdentityPool string, restConfig *rest.Config) (*appsv1.Deployment, error) {
	ctxLog := log.FromContext(ctx)
	ctxLog.Debug("Attempting to generate id porten resource for application", application.Name)

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

	podOpts := pod.PodOpts{
		IstioEnabled: istioEnabled,
	}

	skiperatorContainer := pod.CreateApplicationContainer(application, podOpts)

	var err error

	podVolumes, containerVolumeMounts := volume.GetContainerVolumeMountsAndPodVolumes(application.Spec.FilesFrom)

	if util.IsGCPAuthEnabled(application.Spec.GCP) {
		gcpPodVolume := gcp.GetGCPContainerVolume(workloadIdentityPool, application.Name)
		gcpContainerVolumeMount := gcp.GetGCPContainerVolumeMount()
		gcpEnvVar := gcp.GetGCPEnvVar()

		podVolumes = append(podVolumes, gcpPodVolume)
		containerVolumeMounts = append(containerVolumeMounts, gcpContainerVolumeMount)
		skiperatorContainer.Env = append(skiperatorContainer.Env, gcpEnvVar)
	}

	if idporten.IdportenSpecifiedInSpec(application.Spec.IDPorten) {
		secretName, err := idporten.GetIDPortenSecretName(application.Name)
		if err != nil {
			ctxLog.Error(err, "could not get idporten secret name")
			return nil, err
		}
		podVolumes, containerVolumeMounts = appendDigdiratorSecretVolumeMount(
			&skiperatorContainer,
			containerVolumeMounts,
			podVolumes,
			secretName,
			DefaultDigdiratorIDportenMountPath,
		)
	}

	if maskinporten.MaskinportenSpecifiedInSpec(application.Spec.Maskinporten) {
		secretName, err := maskinporten.GetMaskinportenSecretName(application.Name)
		if err != nil {
			ctxLog.Error(err, "could not get maskinporten secret name")
			return nil, err
		}
		podVolumes, containerVolumeMounts = appendDigdiratorSecretVolumeMount(
			&skiperatorContainer,
			containerVolumeMounts,
			podVolumes,
			secretName,
			DefaultDigdiratorMaskinportenMountPath,
		)
	}

	skiperatorContainer.VolumeMounts = containerVolumeMounts

	var podTemplateLabels map[string]string
	if len(application.Spec.Team) > 0 {
		podTemplateLabels = util.GetPodAppAndTeamSelector(application.Name, application.Spec.Team)
	} else {
		podTemplateLabels = util.GetPodAppSelector(application.Name)
	}

	// Add annotations to pod template, safe-to-evict added due to issues
	// with cluster-autoscaler and unable to evict pods with local volumes
	// https://github.com/kubernetes/autoscaler/blob/master/cluster-autoscaler/FAQ.md
	generatedSpecAnnotations := map[string]string{
		"argocd.argoproj.io/sync-options":                "Prune=false",
		"prometheus.io/scrape":                           "true",
		"cluster-autoscaler.kubernetes.io/safe-to-evict": "true",
	}
	// By specifying port and path annotations, Istio will scrape metrics from the application
	// and merge it together with its own metrics.
	//
	// See
	//  - https://superorbital.io/blog/istio-metrics-merging/
	//  - https://androidexample365.com/an-example-of-how-istio-metrics-merging-works/
	if istioEnabled {
		if application.Spec.Prometheus != nil {
			// If the application has exposed metrics
			generatedSpecAnnotations["prometheus.io/port"] = resolveToPortNumber(application.Spec.Prometheus.Port, application, ctxLog.GetLogger())
			generatedSpecAnnotations["prometheus.io/path"] = application.Spec.Prometheus.Path
		} else {
			// The application doesn't have any custom metrics exposed so we'll disable metrics merging
			// This will ensure that we don't see any messages like this in istio-proxy:
			// "failed scraping application metrics: error scraping http://localhost:80/metrics"
			generatedSpecAnnotations["prometheus.istio.io/merge-metrics"] = "false"
		}
	}

	if application.Spec.PodSettings != nil && len(application.Spec.PodSettings.Annotations) > 0 {
		maps.Copy(generatedSpecAnnotations, application.Spec.PodSettings.Annotations)
	}

	var containers []corev1.Container
	containers = append(containers, skiperatorContainer)

	if util.IsCloudSqlProxyEnabled(application.Spec.GCP) {
		cloudSqlProxyContainer := pod.CreateCloudSqlProxyContainer(application.Spec.GCP.CloudSQLProxy)
		containers = append(containers, cloudSqlProxyContainer)
	}

	podForDeploymentTemplate := corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Labels:      podTemplateLabels,
			Annotations: generatedSpecAnnotations,
		},
		Spec: pod.CreatePodSpec(
			containers,
			podVolumes,
			application.Name,
			application.Spec.Priority,
			util.PointTo(corev1.RestartPolicyAlways),
			application.Spec.PodSettings,
			application.Name,
		),
	}

	deployment.Spec = appsv1.DeploymentSpec{
		Selector: &metav1.LabelSelector{MatchLabels: util.GetPodAppSelector(application.Name)},
		Strategy: appsv1.DeploymentStrategy{
			Type:          appsv1.DeploymentStrategyType(application.Spec.Strategy.Type),
			RollingUpdate: getRollingUpdateStrategy(application.Spec.Strategy.Type),
		},
		Template: corev1.PodTemplateSpec{
			ObjectMeta: podForDeploymentTemplate.ObjectMeta,
			Spec:       podForDeploymentTemplate.Spec,
		},
		RevisionHistoryLimit:    util.PointTo(int32(2)),
		ProgressDeadlineSeconds: util.PointTo(int32(600)),
	}

	// Setting replicas to 0 if manifest has replicas set to 0 or replicas.min/max set to 0
	if shouldScaleToZero(application.Spec.Replicas) {
		deployment.Spec.Replicas = util.PointTo(int32(0))
	}

	if !skiperatorv1alpha1.IsHPAEnabled(application.Spec.Replicas) {
		if replicas, err := skiperatorv1alpha1.GetStaticReplicas(application.Spec.Replicas); err == nil {
			deployment.Spec.Replicas = util.PointTo(int32(replicas))
		} else if replicas, err := skiperatorv1alpha1.GetScalingReplicas(application.Spec.Replicas); err == nil {
			deployment.Spec.Replicas = util.PointTo(int32(replicas.Min))
		} else {
			ctxLog.Error(err, "could not get replicas from application spec")
			return nil, err
		}
	}

	resourceutils.SetCommonAnnotations(&deployment)
	resourceutils.SetApplicationLabels(&deployment, application)

	// add an external link to argocd
	ingresses := application.Spec.Ingresses
	if len(ingresses) > 0 {
		deployment.ObjectMeta.Annotations[AnnotationKeyLinkPrefix] = fmt.Sprintf("https://%s", ingresses[0])
	}

	err = util.ResolveImageTags(ctx, ctxLog.GetLogger(), restConfig, &deployment)
	if err != nil {
		// Exclude dummy image used in tests for decreased verbosity
		if !strings.Contains(err.Error(), "https://index.docker.io/v2/library/image/manifests/latest") {
			ctxLog.Error(err, "could not resolve container image to digest")
		}
		return nil, err
	}

	return &deployment, nil
}

func appendDigdiratorSecretVolumeMount(skiperatorContainer *corev1.Container, volumeMounts []corev1.VolumeMount, volumes []corev1.Volume, secretName string, mountPath string) ([]corev1.Volume, []corev1.VolumeMount) {
	skiperatorContainer.EnvFrom = append(skiperatorContainer.EnvFrom, corev1.EnvFromSource{
		SecretRef: &corev1.SecretEnvSource{
			LocalObjectReference: corev1.LocalObjectReference{
				Name: secretName,
			},
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
				DefaultMode: util.PointTo(int32(420)),
			},
		},
	})

	return volumes, volumeMounts
}

func getRollingUpdateStrategy(updateStrategy string) *appsv1.RollingUpdateDeployment {
	if updateStrategy == "Recreate" {
		return nil
	}

	return &appsv1.RollingUpdateDeployment{
		// Fill with defaults
		MaxUnavailable: &intstr.IntOrString{Type: intstr.String, StrVal: "25%"},
		MaxSurge:       &intstr.IntOrString{Type: intstr.String, StrVal: "25%"},
	}
}

func shouldScaleToZero(jsonReplicas *apiextensionsv1.JSON) bool {
	replicas, err := skiperatorv1alpha1.GetStaticReplicas(jsonReplicas)
	if err == nil && replicas == 0 {
		return true
	}
	replicasStruct, err := skiperatorv1alpha1.GetScalingReplicas(jsonReplicas)
	if err == nil && (replicasStruct.Min == 0 || replicasStruct.Max == 0) {
		return true
	}
	return false
}

func resolveToPortNumber(port intstr.IntOrString, application *skiperatorv1alpha1.Application, ctxLog logr.Logger) string {
	if numericPort := port.IntValue(); numericPort > 0 {
		return fmt.Sprintf("%d", numericPort)
	}

	desiredPortName := port.String()

	if desiredPortName == "main" {
		return fmt.Sprintf("%d", application.Spec.Port)
	}

	for _, p := range application.Spec.AdditionalPorts {
		if p.Name == desiredPortName {
			return fmt.Sprintf("%d", p.Port)
		}
	}

	ctxLog.Error(goerrors.New("port not found"), "could not resolve port name to a port number", "desiredPortName", desiredPortName)
	return desiredPortName
}
