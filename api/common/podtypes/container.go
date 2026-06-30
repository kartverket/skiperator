package podtypes

import corev1 "k8s.io/api/core/v1"

// ContainerSpec describes an extra container to run in the workload's pod
// alongside the main application container.
//
// +kubebuilder:object:generate=true
// +kubebuilder:validation:XValidation:rule="!(self.name in ['cloudsql-proxy', 'istio-proxy', 'istio-validation', 'istio-init'])",message="container name is reserved"
// +kubebuilder:validation:XValidation:rule="!has(self.ingressPort) || (has(self.additionalPorts) && self.additionalPorts.exists(p, p.port == self.ingressPort))",message="ingressPort must be declared in the container's additionalPorts"
type ContainerSpec struct {
	// Name of the container. Must be unique within the pod and must not collide
	// with the application name or a reserved name (e.g. cloudsql-proxy,
	// istio-proxy, istio-validation, istio-init).
	//
	//+kubebuilder:validation:Required
	//+kubebuilder:validation:MinLength=3
	//+kubebuilder:validation:MaxLength=63
	Name string `json:"name"`

	// The container image to run.
	//
	//+kubebuilder:validation:Required
	Image string `json:"image"`

	// Type selects how the container runs:
	//   - "standard" (default): a regular container running alongside the main
	//     container for the lifetime of the pod.
	//   - "init": an init container that starts before the main container and
	//     keeps running for the lifetime of the pod.
	//
	//+kubebuilder:validation:Enum=standard;init
	//+kubebuilder:validation:Optional
	//+kubebuilder:default=standard
	Type string `json:"type,omitempty"`

	// Override the command set in the image.
	//
	//+kubebuilder:validation:Optional
	Command []string `json:"command,omitempty"`

	// Arguments to the container entrypoint.
	//
	//+kubebuilder:validation:Optional
	Args []string `json:"args,omitempty"`

	// Environment variables set inside the container.
	//
	//+kubebuilder:validation:Optional
	Env []corev1.EnvVar `json:"env,omitempty"`

	// Environment variables mounted from ConfigMaps or Secrets. When specified
	// all keys of the resource are assigned as environment variables.
	//
	//+kubebuilder:validation:Optional
	EnvFrom []EnvFrom `json:"envFrom,omitempty"`

	// Files mounted into the container from ConfigMaps, Secrets, PVCs or
	// emptyDirs. The referenced resources are assumed to already exist.
	//
	//+kubebuilder:validation:Optional
	FilesFrom []FilesFrom `json:"filesFrom,omitempty"`

	// Additional ports exposed by the container.
	//
	//+kubebuilder:validation:Optional
	//+kubebuilder:validation:MaxItems=20
	AdditionalPorts []InternalPort `json:"additionalPorts,omitempty"`

	// ResourceRequirements to apply to the container.
	//
	//+kubebuilder:validation:Optional
	Resources *ResourceRequirements `json:"resources,omitempty"`

	// Liveness probe. When provided, path and port are required.
	//
	//+kubebuilder:validation:Optional
	Liveness *Probe `json:"liveness,omitempty"`

	// Readiness probe. When provided, path and port are required.
	//
	//+kubebuilder:validation:Optional
	Readiness *Probe `json:"readiness,omitempty"`

	// Startup probe. When provided, path and port are required.
	//
	//+kubebuilder:validation:Optional
	Startup *Probe `json:"startup,omitempty"`

	// When set, the application's ingress traffic enters the pod through this
	// container instead of the main container: the generated Service keeps its
	// external port (spec.port) but routes its target port to this container's
	// IngressPort. This suits any container that should sit in front of the
	// application and receive incoming traffic first - an auth proxy, an API
	// gateway, a TLS-terminating or rate-limiting proxy, etc. — which then
	// forwards to the application (e.g. it listens on ingressPort and forwards
	// to the app on spec.port via localhost).
	//
	// The IngressPort value must be declared in this container's additionalPorts.
	// At most one extra container may set this, and the value must differ from
	// spec.port.
	//
	//+kubebuilder:validation:Optional
	//+kubebuilder:validation:Minimum=1
	//+kubebuilder:validation:Maximum=65535
	IngressPort *int32 `json:"ingressPort,omitempty"`
}

// The supported values for ContainerSpec.Type.
const (
	// ContainerTypeStandard is a regular container running alongside the main
	// container.
	ContainerTypeStandard = "standard"
	// ContainerTypeInit is an init container, implemented as a native sidecar
	// (init container with restartPolicy: Always).
	ContainerTypeInit = "init"
)
