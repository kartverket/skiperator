package applicationcontroller

import (
	"context"
	goerrors "errors"
	"fmt"
	"github.com/go-logr/logr"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/core"
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
	deploymentLog         = ctrl.Log.WithName("deployment")
	defaultPodAnnotations = map[string]string{
		"argocd.argoproj.io/sync-options": "Prune=false",
		"prometheus.io/scrape":            "true",
	}
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

	podVolumes, containerVolumeMounts := getContainerVolumeMountsAndPodVolumes(application)
	podVolumes, containerVolumeMounts, err = r.appendGCPVolumeMount(application, ctx, &skiperatorContainer, containerVolumeMounts, podVolumes)
	if err != nil {
		r.SetControllerError(ctx, application, controllerName, err)
		return deployment, err
	}
	skiperatorContainer.VolumeMounts = containerVolumeMounts

	labels := util.GetPodAppSelector(application.Name)

	generatedSpecAnnotations := defaultPodAnnotations
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

	deployment.Spec = appsv1.DeploymentSpec{
		Selector: &metav1.LabelSelector{MatchLabels: labels},
		Strategy: appsv1.DeploymentStrategy{
			Type:          appsv1.DeploymentStrategyType(application.Spec.Strategy.Type),
			RollingUpdate: getRollingUpdateStrategy(application.Spec.Strategy.Type),
		},
		Template: corev1.PodTemplateSpec{
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
		},
		RevisionHistoryLimit:    util.PointTo(int32(2)),
		ProgressDeadlineSeconds: util.PointTo(int32(600)),
	}

	// Setting replicas to 0 if manifest has replicas set to 0 or replicas.min/max set to 0
	if shouldScaleToZero(application.Spec.Replicas) {
		deployment.Spec.Replicas = util.PointTo(int32(0))
	}

	if !util.IsHPAEnabled(application.Spec.Replicas) {
		replicas, err := skiperatorv1alpha1.GetReplicasFloat(application.Spec.Replicas)
		if err != nil {
			r.SetControllerError(ctx, application, controllerName, err)
			return deployment, err
		}
		deployment.Spec.Replicas = util.PointTo(int32(replicas))
	}

	r.SetLabelsFromApplication(ctx, &deployment, *application)
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

	deployment := appsv1.Deployment{}
	deploymentDefinition, err := r.defineDeployment(ctx, application)

	err = r.GetClient().Get(ctx, types.NamespacedName{Name: application.Name, Namespace: application.Namespace}, &deployment)
	if err != nil {
		if errors.IsNotFound(err) {
			r.GetRecorder().Eventf(
				application,
				corev1.EventTypeNormal, "NotFound",
				"Deployment resource for application %s not found. Creating deployment",
				application.Name,
			)
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
		if !shouldScaleToZero(application.Spec.Replicas) && util.IsHPAEnabled(application.Spec.Replicas) {
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
						DefaultMode: util.PointTo(int32(420)),
					},
				},
			}
		} else if len(file.Secret) > 0 {
			volume = corev1.Volume{
				Name: file.Secret,
				VolumeSource: corev1.VolumeSource{
					Secret: &corev1.SecretVolumeSource{
						SecretName:  file.Secret,
						DefaultMode: util.PointTo(int32(420)),
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
	if skiperatorv1alpha1.IsReplicasFloat(jsonReplicas) {
		replicas, err := skiperatorv1alpha1.GetReplicasFloat(jsonReplicas)
		if err == nil && replicas == 0 {
			return true
		}
	} else if skiperatorv1alpha1.IsReplicasStruct(jsonReplicas) {
		replicasStruct, err := skiperatorv1alpha1.GetReplicasStruct(jsonReplicas)
		if err == nil && (replicasStruct.Min == 0 || replicasStruct.Max == 0) {
			return true
		}
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
