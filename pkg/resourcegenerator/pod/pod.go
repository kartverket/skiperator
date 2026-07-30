package pod

import (
	"fmt"

	"github.com/kartverket/skiperator/api/common/podtypes"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	skiperatorv1beta1 "github.com/kartverket/skiperator/api/v1beta1"
	"github.com/kartverket/skiperator/internal/config"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/volume"
	"github.com/kartverket/skiperator/pkg/util"
	"github.com/kartverket/skiperator/pkg/util/array"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	DefaultCloudSQLProxyVersion = "2.15.1"
)

type PodOpts struct {
	IstioEnabled     bool
	LocalBuiltImages bool
}

func (po *PodOpts) ImagePullPolicy() corev1.PullPolicy {
	if po.LocalBuiltImages {
		return corev1.PullNever
	}
	return corev1.PullAlways
}

func CreatePodSpec(containers []corev1.Container, volumes []corev1.Volume, serviceAccountName string, priority string,
	policy *corev1.RestartPolicy, podSettings *podtypes.PodSettings, serviceName string) corev1.PodSpec {

	if podSettings == nil {
		podSettings = &podtypes.PodSettings{
			TerminationGracePeriodSeconds: int64(30),
		}
	}

	p := corev1.PodSpec{
		Volumes:                       volumes,
		Containers:                    containers,
		RestartPolicy:                 *policy,
		TerminationGracePeriodSeconds: new(podSettings.TerminationGracePeriodSeconds),
		DNSPolicy:                     corev1.DNSClusterFirst,
		ServiceAccountName:            serviceAccountName,
		DeprecatedServiceAccount:      serviceAccountName,
		NodeName:                      "",
		HostNetwork:                   false,
		HostPID:                       false,
		HostIPC:                       false,
		SecurityContext: &corev1.PodSecurityContext{
			SupplementalGroups: []int64{util.SkiperatorUser},
			FSGroup:            new(util.SkiperatorUser),
			SeccompProfile: &corev1.SeccompProfile{
				Type: corev1.SeccompProfileTypeRuntimeDefault,
			},
		},
		ImagePullSecrets:  []corev1.LocalObjectReference{{Name: "github-auth"}},
		SchedulerName:     corev1.DefaultSchedulerName,
		PriorityClassName: fmt.Sprintf("skip-%s", priority),
	}

	// Allow override per application
	if !podSettings.DisablePodSpreadTopologyConstraints {
		keys := array.TrimmedUniqueStrings(config.GetActiveConfig().TopologyKeys)
		constraints := make([]corev1.TopologySpreadConstraint, 0, len(keys))

		for _, topologyKey := range keys {
			constraints = append(constraints, spreadConstraintForAppAndKey(serviceName, topologyKey))
		}

		p.TopologySpreadConstraints = constraints
	}

	return p
}

// defaultSecurityContext returns the hardened, least-privilege security context
// Skiperator applies to every container it manages (the main application
// container, the CloudSQL proxy and user-provided extra containers). A fresh
// pointer is returned on each call so callers may safely mutate individual
// fields (e.g. the CloudSQL proxy runs as a different UID).
//
// NET_BIND_SERVICE is granted only when allowPrivilegedPorts is true, i.e. the
// container binds a port below 1024. High-port-only containers (e.g. the
// CloudSQL proxy) are left without the capability.
func defaultSecurityContext(allowPrivilegedPorts bool) *corev1.SecurityContext {
	sc := util.LeastPrivilegeContainerSecurityContext.DeepCopy()
	if allowPrivilegedPorts {
		sc.Capabilities.Add = []corev1.Capability{"NET_BIND_SERVICE"}
	}
	return sc
}

// bindsPrivilegedPort reports whether any of the container's declared ports is
// below 1024 and therefore requires NET_BIND_SERVICE. IngressPort is always one
// of AdditionalPorts (enforced by validation), so checking AdditionalPorts is
// sufficient.
func bindsPrivilegedPort(spec podtypes.ContainerSpec) bool {
	for _, p := range spec.AdditionalPorts {
		if p.Port < 1024 {
			return true
		}
	}
	return false
}

func CreateApplicationContainer(application *skiperatorv1alpha1.Application, opts PodOpts) corev1.Container {
	return corev1.Container{
		Name:                     application.Name,
		Image:                    application.Spec.Image,
		ImagePullPolicy:          opts.ImagePullPolicy(),
		Command:                  application.Spec.Command,
		SecurityContext:          defaultSecurityContext(true),
		Ports:                    getContainerPorts(application, opts),
		EnvFrom:                  getEnvFrom(application.Spec.EnvFrom),
		Resources:                getResourceRequirements(application.Spec.Resources),
		Env:                      getEnv(application.Spec.Env),
		ReadinessProbe:           getProbe(application.Spec.Readiness),
		LivenessProbe:            getProbe(application.Spec.Liveness),
		StartupProbe:             getProbe(application.Spec.Startup),
		TerminationMessagePath:   corev1.TerminationMessagePathDefault,
		TerminationMessagePolicy: corev1.TerminationMessageReadFile,
	}
}

// CreateExtraContainers turns the user-provided ExtraContainers specs into pod
// containers. Standard entries are returned as sidecars (appended to
// PodSpec.Containers), while init entries become native sidecars (init
// containers with restartPolicy: Always) returned separately. Volumes derived
// from each container's FilesFrom are returned as-is; callers merge them with
// the pod's existing volumes via AppendUniqueVolumes, which deduplicates by
// name across the whole pod.
func CreateExtraContainers(specs []podtypes.ContainerSpec, opts PodOpts) (sidecars []corev1.Container, initContainers []corev1.Container, volumes []corev1.Volume) {
	for _, spec := range specs {
		containerVolumes := volume.GetPodVolumes(spec.FilesFrom)
		volumeMounts := volume.GetContainerVolumeMounts(spec.FilesFrom)
		volumes = append(volumes, containerVolumes...)

		container := corev1.Container{
			Name:                     spec.Name,
			Image:                    spec.Image,
			ImagePullPolicy:          opts.ImagePullPolicy(),
			Command:                  spec.Command,
			Args:                     spec.Args,
			SecurityContext:          defaultSecurityContext(bindsPrivilegedPort(spec)),
			Ports:                    getInternalContainerPorts(spec.AdditionalPorts),
			EnvFrom:                  getEnvFrom(spec.EnvFrom),
			Env:                      getEnv(spec.Env),
			Resources:                getResourceRequirements(spec.Resources),
			ReadinessProbe:           getProbe(spec.Readiness),
			LivenessProbe:            getProbe(spec.Liveness),
			StartupProbe:             getProbe(spec.Startup),
			VolumeMounts:             volumeMounts,
			TerminationMessagePath:   corev1.TerminationMessagePathDefault,
			TerminationMessagePolicy: corev1.TerminationMessageReadFile,
		}

		if spec.Type == podtypes.ContainerTypeInit {
			// Native sidecar: an init container that keeps running for the
			// lifetime of the pod.
			container.RestartPolicy = new(corev1.ContainerRestartPolicyAlways)
			initContainers = append(initContainers, container)
		} else {
			sidecars = append(sidecars, container)
		}
	}

	return sidecars, initContainers, volumes
}

// AppendUniqueVolumes appends the given volumes to existing, skipping any whose
// name is already present. Used to merge extra-container volumes into the pod's
// volumes without creating duplicate volume names.
func AppendUniqueVolumes(existing []corev1.Volume, toAdd ...corev1.Volume) []corev1.Volume {
	seen := make(map[string]struct{}, len(existing))
	for _, v := range existing {
		seen[v.Name] = struct{}{}
	}
	for _, v := range toAdd {
		if _, ok := seen[v.Name]; ok {
			continue
		}
		seen[v.Name] = struct{}{}
		existing = append(existing, v)
	}
	return existing
}

func getInternalContainerPorts(ports []podtypes.InternalPort) []corev1.ContainerPort {
	var containerPorts []corev1.ContainerPort
	for _, p := range ports {
		containerPorts = append(containerPorts, corev1.ContainerPort{
			Name:          p.Name,
			ContainerPort: p.Port,
			Protocol:      p.Protocol,
		})
	}
	return containerPorts
}

func CreateCloudSqlProxyContainer(cs *podtypes.CloudSQLProxySettings) corev1.Container {
	args := []string{
		cs.ConnectionName,
		"--auto-iam-authn",
		"--structured-logs",
		"--port=5432",
		"--quitquitquit", // Enables admin server at port 9091 and quits when it receives a signal from hahaha
		"--prometheus",   // Enables prometheus metrics at :9090/metrics
	}

	if !cs.PublicIP {
		args = append(args, "--private-ip") // Forces the use of private IP
	}

	if cs.Version == "" {
		cs.Version = DefaultCloudSQLProxyVersion
	}

	// The CloudSQL proxy binds only high ports (5432, 9090, 9091), so it does
	// not need NET_BIND_SERVICE. It also runs as its own dedicated UID/GID.
	googleSC := defaultSecurityContext(false)
	googleSC.RunAsUser = new(int64(200))
	googleSC.RunAsGroup = new(int64(200))

	return corev1.Container{
		Name:            "cloudsql-proxy",
		Image:           "gcr.io/cloud-sql-connectors/cloud-sql-proxy:" + cs.Version,
		ImagePullPolicy: corev1.PullAlways,
		Args:            args,
		SecurityContext: googleSC,
		Resources: corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceMemory: resource.MustParse("64Mi"),
				corev1.ResourceCPU:    resource.MustParse("100m"),
			},
		},
		TerminationMessagePath:   corev1.TerminationMessagePathDefault,
		TerminationMessagePolicy: corev1.TerminationMessageReadFile,
	}
}

func CreateJobContainer(skipJob *skiperatorv1beta1.SKIPJob, volumeMounts []corev1.VolumeMount, envVars []corev1.EnvVar) corev1.Container {
	return corev1.Container{
		Name:                     skipJob.KindPostFixedName(),
		Image:                    skipJob.Spec.Image,
		ImagePullPolicy:          corev1.PullAlways,
		Command:                  skipJob.Spec.Command,
		SecurityContext:          &util.LeastPrivilegeContainerSecurityContext,
		EnvFrom:                  getEnvFrom(skipJob.Spec.EnvFrom),
		Resources:                getResourceRequirements(skipJob.Spec.Resources),
		Env:                      envVars,
		ReadinessProbe:           getProbe(skipJob.Spec.Readiness),
		LivenessProbe:            getProbe(skipJob.Spec.Liveness),
		StartupProbe:             getProbe(skipJob.Spec.Startup),
		TerminationMessagePath:   corev1.TerminationMessagePathDefault,
		TerminationMessagePolicy: corev1.TerminationMessageReadFile,
		VolumeMounts:             volumeMounts,
	}
}

func getProbe(appProbe *podtypes.Probe) *corev1.Probe {
	if appProbe != nil {
		probe := corev1.Probe{
			InitialDelaySeconds: appProbe.InitialDelay,
			TimeoutSeconds:      appProbe.Timeout,
			FailureThreshold:    appProbe.FailureThreshold,
			SuccessThreshold:    appProbe.SuccessThreshold,
			PeriodSeconds:       appProbe.Period,
			ProbeHandler: corev1.ProbeHandler{
				HTTPGet: &corev1.HTTPGetAction{
					Path:   appProbe.Path,
					Port:   appProbe.Port,
					Scheme: corev1.URISchemeHTTP,
				},
			},
		}
		return &probe
	}

	return nil
}

func getResourceRequirements(resources *podtypes.ResourceRequirements) corev1.ResourceRequirements {
	if resources == nil {
		return corev1.ResourceRequirements{}
	}

	return corev1.ResourceRequirements{
		Limits:   (*resources).Limits,
		Requests: (*resources).Requests,
	}
}

func getEnvFrom(envFromApplication []podtypes.EnvFrom) []corev1.EnvFromSource {
	var envFromSource []corev1.EnvFromSource

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

func getEnv(variables []corev1.EnvVar) []corev1.EnvVar {
	for _, variable := range variables {
		if variable.ValueFrom != nil {
			if variable.ValueFrom.FieldRef != nil {
				variable.ValueFrom.FieldRef.APIVersion = "v1"
			}
		}
	}

	return variables
}

func getContainerPorts(application *skiperatorv1alpha1.Application, opts PodOpts) []corev1.ContainerPort {

	containerPorts := []corev1.ContainerPort{
		{
			Name:          "main",
			ContainerPort: int32(application.Spec.Port),
			Protocol:      corev1.ProtocolTCP,
		},
	}

	containerPorts = append(containerPorts, getInternalContainerPorts(application.Spec.AdditionalPorts)...)

	// Expose Prometheus telemetry to Service, so it can be picked up from ServiceMonitor
	if opts.IstioEnabled {
		containerPorts = append(containerPorts, corev1.ContainerPort{
			Name:          util.IstioMetricsPortName.StrVal,
			ContainerPort: util.IstioMetricsPortNumber.IntVal,
			Protocol:      corev1.ProtocolTCP,
		})
	}

	return containerPorts
}

func spreadConstraintForAppAndKey(appName string, key string) corev1.TopologySpreadConstraint {
	return corev1.TopologySpreadConstraint{
		MaxSkew:           1,
		TopologyKey:       key,
		WhenUnsatisfiable: corev1.ScheduleAnyway,
		LabelSelector: &v1.LabelSelector{
			MatchExpressions: []v1.LabelSelectorRequirement{
				{
					Key:      "app",
					Operator: v1.LabelSelectorOpIn,
					Values:   []string{appName},
				},
			},
		},
		// Beta from K8s 1.27, enabled by default
		// See https://medium.com/wise-engineering/avoiding-kubernetes-pod-topology-spread-constraint-pitfalls-d369bb04689e
		MatchLabelKeys: []string{"pod-template-hash"},
	}
}
