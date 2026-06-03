package deployment

import (
	"fmt"
	"strings"

	"github.com/kartverket/skiperator/pkg/reconciliation"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/idporten"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/maskinporten"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/pod"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/resourceutils"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/volume"

	"maps"

	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/gcp"
	"github.com/kartverket/skiperator/pkg/util"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// TODO should clean up
func Generate(r reconciliation.Reconciliation) error {
	ctxLog := r.GetLogger()
	if r.GetType() != reconciliation.ApplicationType {
		err := &reconciliation.SubResourceError{Message: "Unsupported type in deployment resource", WrapErr: fmt.Errorf("unsupported type %s", r.GetType()), Reason: reconciliation.UnsupportedTypeResource}
		return err
	}
	application, ok := r.GetSKIPObject().(*skiperatorv1alpha1.Application)
	if !ok {
		err := &reconciliation.SubResourceError{Message: "Failed to generate deployment resource", WrapErr: fmt.Errorf("failed to cast resource to application"), Reason: reconciliation.InternalError}
		return err
	}

	ctxLog.Debug("Attempting to generate deployment resource for application", "application", application.Name)

	deployment := appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: application.Namespace,
			Name:      application.Name,
		},
	}

	podOpts := pod.PodOpts{
		IstioEnabled:     r.IsIstioEnabled(),
		LocalBuiltImages: r.GetSkiperatorConfig().EnableLocallyBuiltImages,
	}

	skiperatorContainer := pod.CreateApplicationContainer(application, podOpts)

	podVolumes, containerVolumeMounts := volume.GetContainerVolumeMountsAndPodVolumes(application.Spec.FilesFrom)

	if util.IsGCPAuthEnabled(application.Spec.GCP) {
		gcpPodVolume := gcp.GetGCPContainerVolume(r.GetSkiperatorConfig().GCPWorkloadIdentityPool, application.Name)
		gcpContainerVolumeMount := gcp.GetGCPContainerVolumeMount()
		gcpEnvVar := gcp.GetGCPEnvVar()

		podVolumes = append(podVolumes, gcpPodVolume)
		containerVolumeMounts = append(containerVolumeMounts, gcpContainerVolumeMount)
		skiperatorContainer.Env = append(skiperatorContainer.Env, gcpEnvVar)
	}

	if idporten.IdportenSpecifiedInSpec(application.Spec.IDPorten) {
		secretName, err := idporten.GetIDPortenSecretName(application.Name)
		if err != nil {
			err := &reconciliation.SubResourceError{Message: "Failed to get idporten secret name", WrapErr: err, Reason: reconciliation.ResourceDependencyNotFound}
			return err
		}
		podVolumes, containerVolumeMounts = volume.AppendDigdiratorSecret(
			&skiperatorContainer,
			containerVolumeMounts,
			podVolumes,
			secretName,
			volume.DefaultDigdiratorIDportenMountPath,
		)
	}

	if maskinporten.MaskinportenSpecifiedInSpec(application.Spec.Maskinporten) {
		secretName, err := maskinporten.GetMaskinportenSecretName(application.Name)
		if err != nil {
			err := &reconciliation.SubResourceError{Message: "Failed to get maskinporten secret name", WrapErr: err, Reason: reconciliation.ResourceDependencyNotFound}
			return err
		}
		podVolumes, containerVolumeMounts = volume.AppendDigdiratorSecret(
			&skiperatorContainer,
			containerVolumeMounts,
			podVolumes,
			secretName,
			volume.DefaultDigdiratorMaskinportenMountPath,
		)
	}

	skiperatorContainer.VolumeMounts = containerVolumeMounts

	var podTemplateLabels map[string]string
	if len(application.Spec.Team) > 0 {
		podTemplateLabels = util.GetPodAppAndTeamSelector(application.Name, application.Spec.Team)
	} else {
		podTemplateLabels = util.GetPodAppSelector(application.Name)
	}
	podTemplateLabels["app.kubernetes.io/version"] = resourceutils.HumanReadableVersion(&ctxLog, application.Spec.Image)

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
	if r.IsIstioEnabled() {
		if application.Spec.Prometheus != nil {
			// If the application has exposed metrics
			generatedSpecAnnotations["prometheus.io/port"] = application.ResolvePortNumber(application.Spec.Prometheus.Port, ctxLog.GetLogger())
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

	containers := []corev1.Container{skiperatorContainer}

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
			new(corev1.RestartPolicyAlways),
			application.Spec.PodSettings,
			application.Name,
		),
	}

	//we need to set the pod labels like this as its a template, not a resource.
	//TODO: figure out a smoother solution?
	resourceutils.SetApplicationLabels(&podForDeploymentTemplate, application)
	resourceutils.SetCommonAnnotations(&podForDeploymentTemplate)

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
		RevisionHistoryLimit:    new(int32(2)),
		ProgressDeadlineSeconds: new(int32(600)),
	}

	// Setting replicas to 0 if manifest has replicas set to 0 or replicas.min/max set to 0
	if resourceutils.ShouldScaleToZero(application.Spec.Replicas) {
		deployment.Spec.Replicas = new(int32(0))
	}

	if !skiperatorv1alpha1.IsHPAEnabled(application.Spec.Replicas) {
		if replicas, err := skiperatorv1alpha1.GetStaticReplicas(application.Spec.Replicas); err == nil {
			deployment.Spec.Replicas = new(int32(replicas))
		} else if replicas, err := skiperatorv1alpha1.GetScalingReplicas(application.Spec.Replicas); err == nil {
			deployment.Spec.Replicas = new(int32(replicas.Min))
		} else {
			err := &reconciliation.SubResourceError{Message: "Failed to get replicas from application spec", WrapErr: err, Reason: reconciliation.InternalError}
			return err
		}
	}

	// add an external link to argocd
	ingresses := application.Spec.Ingresses
	if deployment.Annotations == nil {
		deployment.Annotations = make(map[string]string)
	}

	if len(ingresses) > 0 {
		deployment.Annotations[resourceutils.AnnotationKeyLinkPrefix] = fmt.Sprintf("https://%s", ingresses[0])
	}

	if !podOpts.LocalBuiltImages {
		err := util.ResolveImageTags(r.GetCtx(), ctxLog.GetLogger(), r.GetRestConfig(), &deployment)
		if err != nil {
			//TODO fix this
			// Exclude dummy image used in tests for decreased verbosity
			if !strings.Contains(err.Error(), "https://index.docker.io/v2/library/image/manifests/latest") {
				err := &reconciliation.SubResourceError{Message: "Could not resolve container image to digest", WrapErr: err, Reason: reconciliation.ContainerImageNotFound}
				return err
			}
		}
	}

	r.AddResource(&deployment)

	ctxLog.Debug("successfully created deployment resource")
	return nil
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
