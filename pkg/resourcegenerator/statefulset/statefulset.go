package statefulset

import (
	"fmt"
	"maps"
	"strings"

	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/reconciliation"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/gcp"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/idporten"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/maskinporten"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/pod"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/resourceutils"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/volume"
	"github.com/kartverket/skiperator/pkg/util"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	AnnotationKeyLinkPrefix                = "link.argocd.argoproj.io/external-link"
	DefaultDigdiratorMaskinportenMountPath = "/var/run/secrets/skip/maskinporten"
	DefaultDigdiratorIDportenMountPath     = "/var/run/secrets/skip/idporten"

	HeadlessServiceSuffix = "-headless"
)

func HeadlessServiceName(appName string) string {
	return appName + HeadlessServiceSuffix
}

func Generate(r reconciliation.Reconciliation) error {
	ctxLog := r.GetLogger()
	if r.GetType() != reconciliation.ApplicationType {
		return &reconciliation.SubResourceError{Message: "Unsupported type in statefulset resource", WrapErr: fmt.Errorf("unsupported type %s", r.GetType()), Reason: reconciliation.UnsupportedTypeResource}
	}
	application, ok := r.GetSKIPObject().(*skiperatorv1alpha1.Application)
	if !ok {
		return &reconciliation.SubResourceError{Message: "Failed to generate statefulset resource", WrapErr: fmt.Errorf("failed to cast resource to application"), Reason: reconciliation.InternalError}
	}

	ctxLog.Debug("Attempting to generate statefulset resource for application", "application", application.Name)

	sts := appsv1.StatefulSet{
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
			return &reconciliation.SubResourceError{Message: "Failed to get idporten secret name", WrapErr: err, Reason: reconciliation.ResourceDependencyNotFound}
		}
		podVolumes, containerVolumeMounts = volume.AppendDigdiratorSecret(
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
			return &reconciliation.SubResourceError{Message: "Failed to get maskinporten secret name", WrapErr: err, Reason: reconciliation.ResourceDependencyNotFound}
		}
		podVolumes, containerVolumeMounts = volume.AppendDigdiratorSecret(
			&skiperatorContainer,
			containerVolumeMounts,
			podVolumes,
			secretName,
			DefaultDigdiratorMaskinportenMountPath,
		)
	}

	var vctTemplates []corev1.PersistentVolumeClaim
	for _, vct := range application.Spec.VolumeClaimTemplates {
		vctTemplates = append(vctTemplates, corev1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name:        vct.Name,
				Labels:      vct.Labels,
				Annotations: vct.Annotations,
			},
			Spec: vct.Spec,
		})
		containerVolumeMounts = append(containerVolumeMounts, corev1.VolumeMount{
			Name:      vct.Name,
			MountPath: vct.MountPath,
			SubPath:   vct.SubPath,
		})
	}

	skiperatorContainer.VolumeMounts = containerVolumeMounts

	var podTemplateLabels map[string]string
	if len(application.Spec.Team) > 0 {
		podTemplateLabels = util.GetPodAppAndTeamSelector(application.Name, application.Spec.Team)
	} else {
		podTemplateLabels = util.GetPodAppSelector(application.Name)
	}
	podTemplateLabels["app.kubernetes.io/version"] = resourceutils.HumanReadableVersion(&ctxLog, application.Spec.Image)

	generatedSpecAnnotations := map[string]string{
		"argocd.argoproj.io/sync-options":                "Prune=false",
		"prometheus.io/scrape":                           "true",
		"cluster-autoscaler.kubernetes.io/safe-to-evict": "true",
	}
	if r.IsIstioEnabled() {
		if application.Spec.Prometheus != nil {
			generatedSpecAnnotations["prometheus.io/port"] = application.ResolvePortNumber(application.Spec.Prometheus.Port, ctxLog.GetLogger())
			generatedSpecAnnotations["prometheus.io/path"] = application.Spec.Prometheus.Path
		} else {
			generatedSpecAnnotations["prometheus.istio.io/merge-metrics"] = "false"
		}
	}

	if application.Spec.PodSettings != nil && len(application.Spec.PodSettings.Annotations) > 0 {
		maps.Copy(generatedSpecAnnotations, application.Spec.PodSettings.Annotations)
	}

	containers := []corev1.Container{skiperatorContainer}

	if util.IsCloudSqlProxyEnabled(application.Spec.GCP) {
		containers = append(containers, pod.CreateCloudSqlProxyContainer(application.Spec.GCP.CloudSQLProxy))
	}

	podTemplate := corev1.Pod{
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

	resourceutils.SetApplicationLabels(&podTemplate, application)
	resourceutils.SetCommonAnnotations(&podTemplate)

	sts.Spec = appsv1.StatefulSetSpec{
		Selector:    &metav1.LabelSelector{MatchLabels: util.GetPodAppSelector(application.Name)},
		ServiceName: HeadlessServiceName(application.Name),
		Template: corev1.PodTemplateSpec{
			ObjectMeta: podTemplate.ObjectMeta,
			Spec:       podTemplate.Spec,
		},
		VolumeClaimTemplates:                 vctTemplates,
		PodManagementPolicy:                  appsv1.PodManagementPolicyType(application.Spec.PodManagementPolicy),
		UpdateStrategy:                       buildUpdateStrategy(application.Spec.Partition),
		PersistentVolumeClaimRetentionPolicy: buildPVCRetentionPolicy(application.Spec.PVCRetentionWhenDeleted, application.Spec.PVCRetentionWhenScaled),
		RevisionHistoryLimit:                 new(int32(2)),
	}

	replicas, err := skiperatorv1alpha1.GetStaticReplicas(application.Spec.Replicas)
	if err == nil {
		sts.Spec.Replicas = new(int32(replicas))
	} else {
		return &reconciliation.SubResourceError{Message: "Failed to get replicas from application spec (stateful workloads require static replicas)", WrapErr: err, Reason: reconciliation.InternalError}
	}

	if sts.Annotations == nil {
		sts.Annotations = make(map[string]string)
	}

	ingresses := application.Spec.Ingresses
	if len(ingresses) > 0 {
		sts.Annotations[AnnotationKeyLinkPrefix] = fmt.Sprintf("https://%s", ingresses[0])
	}

	if !podOpts.LocalBuiltImages {
		err := util.ResolveImageTags(r.GetCtx(), ctxLog.GetLogger(), r.GetRestConfig(), &sts)
		if err != nil {
			if !strings.Contains(err.Error(), "https://index.docker.io/v2/library/image/manifests/latest") {
				return &reconciliation.SubResourceError{Message: "Could not resolve container image to digest", WrapErr: err, Reason: reconciliation.ContainerImageNotFound}
			}
		}
	}

	r.AddResource(&sts)

	ctxLog.Debug("successfully created statefulset resource")
	return nil
}

func buildUpdateStrategy(partition *int32) appsv1.StatefulSetUpdateStrategy {
	strategy := appsv1.StatefulSetUpdateStrategy{
		Type: appsv1.RollingUpdateStatefulSetStrategyType,
	}
	if partition != nil {
		strategy.RollingUpdate = &appsv1.RollingUpdateStatefulSetStrategy{
			Partition: partition,
		}
	}
	return strategy
}

func buildPVCRetentionPolicy(whenDeleted, whenScaled string) *appsv1.StatefulSetPersistentVolumeClaimRetentionPolicy {
	if whenDeleted == "" && whenScaled == "" {
		return nil
	}
	policy := &appsv1.StatefulSetPersistentVolumeClaimRetentionPolicy{}
	if whenDeleted != "" {
		policy.WhenDeleted = appsv1.PersistentVolumeClaimRetentionPolicyType(whenDeleted)
	}
	if whenScaled != "" {
		policy.WhenScaled = appsv1.PersistentVolumeClaimRetentionPolicyType(whenScaled)
	}
	return policy
}
