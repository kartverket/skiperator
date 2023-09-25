package applicationcontroller

import (
	"context"
	goerrors "errors"
	"fmt"
	"github.com/go-logr/logr"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/core"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/gcp"
	"github.com/kartverket/skiperator/pkg/util"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"strings"
)

const (
	AnnotationKeyLinkPrefix = "link.argocd.argoproj.io/external-link"
)

var (
	deploymentLog = ctrl.Log.WithName("deployment")
)

func (r *ApplicationReconciler) defineDeployment(ctx context.Context, application *skiperatorv1alpha1.Application) (appsv1.Deployment, error) {
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

	podOpts := core.PodOpts{
		IstioEnabled: r.IsIstioEnabledForNamespace(ctx, application.Namespace),
	}

	skiperatorContainer := core.CreateApplicationContainer(application, podOpts)

	var err error

	podVolumes, containerVolumeMounts := core.GetContainerVolumeMountsAndPodVolumes(application.Spec.FilesFrom)

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
			return deployment, err
		}

		gcpPodVolume := gcp.GetGCPContainerVolume(gcpIdentityConfigMap.Data["workloadIdentityPool"], application.Name)
		gcpContainerVolumeMount := gcp.GetGCPContainerVolumeMount()
		gcpEnvVar := gcp.GetGCPEnvVar()

		podVolumes = append(podVolumes, gcpPodVolume)
		containerVolumeMounts = append(containerVolumeMounts, gcpContainerVolumeMount)
		skiperatorContainer.Env = append(skiperatorContainer.Env, gcpEnvVar)
	}

	skiperatorContainer.VolumeMounts = containerVolumeMounts

	labels := util.GetPodAppSelector(application.Name)

	generatedSpecAnnotations := map[string]string{
		"argocd.argoproj.io/sync-options": "Prune=false",
		"prometheus.io/scrape":            "true",
	}
	// By specifying port and path annotations, Istio will scrape metrics from the application
	// and merge it together with its own metrics.
	//
	// See
	//  - https://superorbital.io/blog/istio-metrics-merging/
	//  - https://androidexample365.com/an-example-of-how-istio-metrics-merging-works/
	istioEnabled := r.IsIstioEnabledForNamespace(ctx, application.Namespace)
	if istioEnabled && application.Spec.Prometheus != nil {
		generatedSpecAnnotations["prometheus.io/port"] = resolveToPortNumber(application.Spec.Prometheus.Port, application)
		generatedSpecAnnotations["prometheus.io/path"] = application.Spec.Prometheus.Path
	}

	podForDeploymentTemplate := corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Labels:      labels,
			Annotations: generatedSpecAnnotations,
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
			// The resulting kubernetes object includes the ServiceAccount field, and thus it's required in order
			// to not create a diff for the hash of existing and wanted spec
			DeprecatedServiceAccount:      application.Name,
			Volumes:                       podVolumes,
			PriorityClassName:             fmt.Sprintf("skip-%s", application.Spec.Priority),
			RestartPolicy:                 corev1.RestartPolicyAlways,
			TerminationGracePeriodSeconds: util.PointTo(int64(corev1.DefaultTerminationGracePeriodSeconds)),
			DNSPolicy:                     corev1.DNSClusterFirst,
			SchedulerName:                 corev1.DefaultSchedulerName,
		},
	}

	r.SetLabelsFromApplication(&podForDeploymentTemplate, *application)

	deployment.Spec = appsv1.DeploymentSpec{
		Selector: &metav1.LabelSelector{MatchLabels: labels},
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
			r.SetControllerError(ctx, application, controllerName, err)
			return deployment, err
		}
	}

	r.SetLabelsFromApplication(&deployment, *application)
	util.SetCommonAnnotations(&deployment)

	// add an external link to argocd
	ingresses := application.Spec.Ingresses
	if len(ingresses) > 0 {
		deployment.ObjectMeta.Annotations[AnnotationKeyLinkPrefix] = fmt.Sprintf("https://%s", ingresses[0])
	}

	// Set application as owner of the deployment
	err = ctrlutil.SetControllerReference(application, &deployment, r.GetScheme())
	if err != nil {
		r.SetControllerError(ctx, application, controllerName, err)
		return deployment, err
	}

	return *r.resolveDigest(ctx, &deployment), nil
}

func (r *ApplicationReconciler) reconcileDeployment(ctx context.Context, application *skiperatorv1alpha1.Application) (reconcile.Result, error) {
	controllerName := "Deployment"
	r.SetControllerProgressing(ctx, application, controllerName)

	deployment := appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      application.Name,
			Namespace: application.Namespace,
		},
	}
	deploymentDefinition, err := r.defineDeployment(ctx, application)

	shouldReconcile, err := r.ShouldReconcile(ctx, &deployment)
	if err != nil {
		r.SetControllerFinishedOutcome(ctx, application, controllerName, err)
		return reconcile.Result{}, err
	}

	if !shouldReconcile {
		r.SetControllerFinishedOutcome(ctx, application, controllerName, err)
		return reconcile.Result{}, nil
	}

	err = r.GetClient().Get(ctx, client.ObjectKeyFromObject(&deployment), &deployment)
	if err != nil {
		if errors.IsNotFound(err) {
			r.EmitNormalEvent(application, "NotFound", fmt.Sprintf("deployment resource for application %s not found, creating deployment", application.Name))
			err = r.GetClient().Create(ctx, &deploymentDefinition)
			if err != nil {
				r.SetControllerError(ctx, application, controllerName, err)
				return reconcile.Result{}, err
			}
		} else {
			r.SetControllerError(ctx, application, controllerName, err)
			return reconcile.Result{}, err
		}
	} else {
		if !shouldScaleToZero(application.Spec.Replicas) && skiperatorv1alpha1.IsHPAEnabled(application.Spec.Replicas) {
			// Ignore replicas set by HPA when checking diff
			if int32(*deployment.Spec.Replicas) > 0 {
				deployment.Spec.Replicas = nil
			}
		}
		deployment = *r.resolveDigest(ctx, &deployment)

		deploymentHash := util.GetHashForStructs([]interface{}{
			&deployment.Spec,
			&deployment.Labels,
		})
		deploymentDefinitionHash := util.GetHashForStructs([]interface{}{
			&deploymentDefinition.Spec,
			&deploymentDefinition.Labels,
		})

		if deploymentHash != deploymentDefinitionHash {
			patch := client.MergeFrom(deployment.DeepCopy())
			err = r.GetClient().Patch(ctx, &deploymentDefinition, patch)
			if err != nil {
				r.SetControllerError(ctx, application, controllerName, err)
				return reconcile.Result{}, err
			}
		}
	}

	r.SetControllerFinishedOutcome(ctx, application, controllerName, err)

	return reconcile.Result{}, err
}

func (r *ApplicationReconciler) resolveDigest(ctx context.Context, input *appsv1.Deployment) *appsv1.Deployment {
	res, err := util.ResolveImageTags(ctx, logr.Discard(), r.GetRestConfig(), input)
	if err != nil {
		// Exclude dummy image used in tests for decreased verbosity
		if !strings.Contains(err.Error(), "https://index.docker.io/v2/library/image/manifests/latest") {
			deploymentLog.Error(err, "could not resolve container image to digest")
		}
		return input
	}
	// FIXME: Consider setting imagePullPolicy=IfNotPresent when the image has been resolved to
	// a digest in order to reduce registry usage and spin-up times.
	return res
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

func resolveToPortNumber(port intstr.IntOrString, application *skiperatorv1alpha1.Application) string {
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

	deploymentLog.Error(goerrors.New("port not found"), "could not resolve port name to a port number", "desiredPortName", desiredPortName)
	return desiredPortName
}
