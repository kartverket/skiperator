package v1alpha1

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/kartverket/skiperator/api/common"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
type ApplicationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Application `json:"items"`
}

// Application
//
// Root object for Application resource. An application resource is a resource for easily managing a Dockerized container within the context of a Kartverket cluster.
// This allows product teams to avoid the need to set up networking on the cluster, as well as a lot of out of the box security features.
//
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName="app"
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.summary.status`
// +kubebuilder:printcolumn:name="AccessPolicies",type=string,JSONPath=`.status.accessPolicies`
type Application struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ApplicationSpec  `json:"spec,omitempty"`
	Status SkiperatorStatus `json:"status,omitempty"`
}

// +kubebuilder:object:generate=true
type ApplicationSpec struct {
	// The image the application will run. This image will be added to a Deployment resource
	//
	//+kubebuilder:validation:Required
	Image string `json:"image"`

	// The port the deployment exposes
	//
	//+kubebuilder:validation:Required
	Port int `json:"port"`

	// Protocol that the application speaks.
	//
	//+kubebuilder:validation:Enum=http;tcp;udp
	//+kubebuilder:default=http
	AppProtocol string `json:"appProtocol,omitempty"`

	// Any external hostnames that route to this application. Using a skip.statkart.no-address
	// will make the application reachable for kartverket-clients (internal), other addresses
	// make the app reachable on the internet. Note that other addresses than skip.statkart.no
	// (also known as pretty hostnames) requires additional DNS setup.
	// The below hostnames will also have TLS certificates issued and be reachable on both
	// HTTP and HTTPS.
	//
	// Ingresses must be lowercase, contain no spaces, be a non-empty string, and have a hostname/domain separated by a period
	// They can optionally be suffixed with a plus and name of a custom TLS secret located in the istio-gateways namespace.
	// E.g. "foo.atkv3-dev.kartverket-intern.cloud+env-wildcard-cert"
	//+kubebuilder:validation:Optional
	Ingresses []string `json:"ingresses,omitempty"`

	// An optional priority. Supported values are 'low', 'medium' and 'high'.
	// The default value is 'medium'.
	//
	// Most workloads should not have to specify this field. If you think you
	// do, please consult with SKIP beforehand.
	//
	//+kubebuilder:validation:Enum=low;medium;high
	//+kubebuilder:default=medium
	Priority string `json:"priority,omitempty"`

	// Team specifies the team who owns this particular app.
	// Usually sourced from the namespace label.
	//
	//+kubebuilder:validation:Optional
	Team string `json:"team,omitempty"`

	// Override the command set in the Dockerfile. Usually only used when debugging
	// or running third-party containers where you don't have control over the Dockerfile
	//
	//+kubebuilder:validation:Optional
	Command []string `json:"command,omitempty"`

	// ResourceRequirements to apply to the deployment. It's common to set some of these to
	// prevent the app from swelling in resource usage and consuming all the
	// resources of other apps on the cluster.
	//
	//+kubebuilder:validation:Optional
	Resources *ResourceRequirements `json:"resources,omitempty"`

	// The number of replicas can either be specified as a static number as follows:
	//
	// 	replicas: 2
	//
	// Or by specifying a range between min and max to enable HorizontalPodAutoscaling.
	// The default value for replicas is:
	// 	replicas:
	// 		min: 2
	// 		max: 5
	// 		targetCpuUtilization: 80
	//      targetMemoryUtilization: 80
	// Using autoscaling is the recommended configuration for replicas.
	//+kubebuilder:validation:Optional
	Replicas *apiextensionsv1.JSON `json:"replicas,omitempty"`

	// Defines an alternative strategy for the Kubernetes deployment. This is useful when
	// the default strategy, RollingUpdate, is not usable. Setting type to
	// Recreate will take down all the pods before starting new pods, whereas the
	// default of RollingUpdate will try to start the new pods before taking down the
	// old ones.
	//
	// Valid values are: RollingUpdate, Recreate. Default is RollingUpdate
	//
	//+kubebuilder:validation:Optional
	Strategy Strategy `json:"strategy,omitempty"`

	// Environment variables that will be set inside the Deployment's Pod. See https://pkg.go.dev/k8s.io/api/core/v1#EnvVar for examples.
	//
	//+kubebuilder:validation:Optional
	Env []corev1.EnvVar `json:"env,omitempty"`

	// Environment variables mounted from files. When specified all the keys of the
	// resource will be assigned as environment variables. Supports both configmaps
	// and secrets.
	//
	// For mounting as files see FilesFrom.
	//
	//+kubebuilder:validation:Optional
	EnvFrom []EnvFrom `json:"envFrom,omitempty"`

	// Mounting volumes into the Deployment are done using the FilesFrom argument
	//
	// FilesFrom supports ConfigMaps, Secrets and PVCs. The Application resource
	// assumes these have already been created by you, and will fail if this is not the case.
	//
	// For mounting environment variables see EnvFrom.
	//
	//+kubebuilder:validation:Optional
	FilesFrom []FilesFrom `json:"filesFrom,omitempty"`

	// An optional list of extra port to expose on a pod level basis,
	// for example so Instana or other APM tools can reach it
	//
	//+kubebuilder:validation:Optional
	AdditionalPorts []InternalPort `json:"additionalPorts,omitempty"`
	// Liveness probes define a resource that returns 200 OK when the app is running
	// as intended. Returning a non-200 code will make kubernetes restart the app.
	// Liveness is optional, but when provided, path and port are required
	//
	// See Probe for structure definition.
	//
	//+kubebuilder:validation:Optional
	Liveness *Probe `json:"liveness,omitempty"`

	// Readiness probes define a resource that returns 200 OK when the app is running
	// as intended. Kubernetes will wait until the resource returns 200 OK before
	// marking the pod as Running and progressing with the deployment strategy.
	// Readiness is optional, but when provided, path and port are required
	//
	//+kubebuilder:validation:Optional
	Readiness *Probe `json:"readiness,omitempty"`

	// Kubernetes uses startup probes to know when a container application has started.
	// If such a probe is configured, it disables liveness and readiness checks until it
	// succeeds, making sure those probes don't interfere with the application startup.
	// This can be used to adopt liveness checks on slow starting containers, avoiding them
	// getting killed by Kubernetes before they are up and running.
	// Startup is optional, but when provided, path and port are required
	//
	//+kubebuilder:validation:Optional
	Startup *Probe `json:"startup,omitempty"`

	// Settings for Maskinporten integration with Digitaliseringsdirektoratet
	//
	//+kubebuilder:validation:Optional
	Maskinporten *Maskinporten `json:"maskinporten,omitempty"`

	// Settings for IDPorten integration with Digitaliseringsdirektoratet
	//
	//+kubebuilder:validation:Optional
	IDPorten *IDPorten `json:"idporten,omitempty"`

	// Optional settings for how Prometheus compatible metrics should be scraped.
	//
	//+kubebuilder:validation:Optional
	Prometheus *PrometheusConfig `json:"prometheus,omitempty"`

	// Controls whether the application will automatically redirect all HTTP calls to HTTPS via the istio VirtualService.
	// This redirect does not happen on the route /.well-known/acme-challenge/, as the ACME challenge can only be done on port 80.
	//
	//+kubebuilder:validation:Optional
	//+kubebuilder:default:=true
	RedirectToHTTPS *bool `json:"redirectToHTTPS,omitempty"`

	// Whether to enable automatic Pod Disruption Budget creation for this application.
	//
	//+kubebuilder:validation:Optional
	//+kubebuilder:default=true
	EnablePDB *bool `json:"enablePDB,omitempty"`

	// The root AccessPolicy for managing zero trust access to your Application. See AccessPolicy for more information.
	//
	//+kubebuilder:validation:Optional
	AccessPolicy *AccessPolicy `json:"accessPolicy,omitempty"`

	// GCP is used to configure Google Cloud Platform specific settings for the application.
	//
	//+kubebuilder:validation:Optional
	GCP *GCP `json:"gcp,omitempty"`

	// Labels can be used if you want every resource created by your application to
	// have the same labels, including your application. This could for example be useful for
	// metrics, where a certain label and the corresponding resources liveliness can be combined.
	// Any amount of labels can be added as wanted, and they will all cascade down to all resources.
	//
	//+kubebuilder:validation:Optional
	Labels map[string]string `json:"labels,omitempty"`

	// ResourceLabels can be used if you want to add a label to a specific resources created by
	// the application. One such label could for example be set on a Deployment, such that
	// the deployment avoids certain rules from Gatekeeper, or similar. Any amount of labels may be added per ResourceLabels item.
	//
	//+kubebuilder:validation:Optional
	ResourceLabels map[string]map[string]string `json:"resourceLabels,omitempty"`

	// Used for allow listing certain default blocked endpoints, such as /actuator/ end points
	//
	//+kubebuilder:validation:Optional
	AuthorizationSettings *AuthorizationSettings `json:"authorizationSettings,omitempty"`

	// PodSettings are used to apply specific settings to the Pod Template used by Skiperator to create Deployments. This allows you to set
	// things like annotations on the Pod to change the behaviour of sidecars, and set relevant Pod options such as TerminationGracePeriodSeconds.
	//
	//+kubebuilder:validation:Optional
	PodSettings *PodSettings `json:"podSettings,omitempty"`

	// IstioSettings are used to configure istio specific resources such as telemetry. Currently, adjusting sampling
	// interval for tracing is the only supported option.
	// By default, tracing is enabled with a random sampling percentage of 10%.
	//
	//+kubebuilder:validation:Optional
	//+kubebuilder:default:={telemetry: {tracing: {{randomSamplingPercentage: 10}}}}
	IstioSettings *IstioSettingsApplication `json:"istioSettings,omitempty"`
}

// AuthorizationSettings Settings for overriding the default deny of all actuator endpoints. AllowAll will allow any
// endpoint to be exposed. Use AllowList to only allow specific endpoints.
//
// Please be aware that HTTP endpoints, such as actuator, may expose information about your application which you do not want to expose.
// Before allow listing HTTP endpoints, make note of what these endpoints will expose, especially if your application is served via an external ingress.
//
// +kubebuilder:object:generate=true
type AuthorizationSettings struct {
	// Allows all endpoints by not creating an AuthorizationPolicy, and ignores the content of AllowList.
	// If field is false, the contents of AllowList will be used instead if AllowList is set.
	//
	//+kubebuilder:validation:Optional
	//+kubebuilder:default:=false
	AllowAll bool `json:"allowAll,omitempty"`

	// Allows specific endpoints. Common endpoints one might want to allow include /actuator/health, /actuator/startup, /actuator/info.
	//
	// Note that endpoints are matched specifically on the input, so if you allow /actuator/health, you will *not* allow /actuator/health/
	//
	//+kubebuilder:validation:Optional
	AllowList []string `json:"allowList,omitempty"`
}

// +kubebuilder:object:generate=true
type Replicas struct {
	// Min represents the minimum number of replicas when load is low.
	// Note that the SKIP team recommends that you set this to at least two, but this is only required for production.
	//
	//+kubebuilder:validation:Required
	Min uint `json:"min"`

	// Max represents the maximum number of replicas the deployment is allowed to scale to
	//
	//+kubebuilder:validation:Optional
	Max uint `json:"max,omitempty"`

	// When the average CPU utilization across all pods crosses this threshold another replica is started, up to a maximum of Max
	//
	// TargetCpuUtilization is an integer representing a percentage.
	//
	//+kubebuilder:default:=80
	//+kubebuilder:validation:Optional
	TargetCpuUtilization uint `json:"targetCpuUtilization,omitempty"`
	// When the average Memory utilization across all pods crosses this threshold another replica is started, up to a maximum of Max
	//
	// TargetMemoryUtilization is an integer representing a percentage.
	//
	//+kubebuilder:default:=80
	//+kubebuilder:validation:Optional
	TargetMemoryUtilization uint `json:"targetMemoryUtilization,omitempty"`
}

// Strategy
//
// Object representing a Kubernetes deployment strategy. Currently only contains a Type object,
// could probably be omitted in favour of directly using the Type.
//
// +kubebuilder:object:generate=true
type Strategy struct {
	// Valid values are: RollingUpdate, Recreate. Default is RollingUpdate
	//
	//+kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum=RollingUpdate;Recreate
	// +kubebuilder:default=RollingUpdate
	Type string `json:"type,omitempty"`
}

func NewDefaultReplicas() Replicas {
	return Replicas{
		Min:                  2,
		Max:                  5,
	}
}

func (replicas *Replicas) ApplyDefaultUtilization() {
	// set default on both target values if none are set
	if replicas.TargetCpuUtilization == 0 && replicas.TargetMemoryUtilization == 0 {
		replicas.TargetCpuUtilization = 80
		replicas.TargetMemoryUtilization = 80
	}
}

func MarshalledReplicas(replicas interface{}) *apiextensionsv1.JSON {
	replicasJson := &apiextensionsv1.JSON{}
	var err error

	replicasJson.Raw, err = json.Marshal(replicas)
	if err == nil {
		return replicasJson
	}

	return nil
}

func GetStaticReplicas(jsonReplicas *apiextensionsv1.JSON) (uint, error) {
	var result uint
	err := json.Unmarshal(jsonReplicas.Raw, &result)

	return result, err
}

func GetScalingReplicas(jsonReplicas *apiextensionsv1.JSON) (Replicas, error) {
	result := NewDefaultReplicas()
	err := json.Unmarshal(jsonReplicas.Raw, &result)
	result.ApplyDefaultUtilization()
	return result, err
}

func IsHPAEnabled(jsonReplicas *apiextensionsv1.JSON) bool {
	replicas, err := GetScalingReplicas(jsonReplicas)
	if err == nil &&
		replicas.Min > 0 &&
		replicas.Min < replicas.Max {
		return true
	}
	return false
}

func (a *Application) FillDefaultsSpec() {
	if a.Spec.Replicas == nil {
		defaultReplicas := NewDefaultReplicas()
		defaultReplicas.ApplyDefaultUtilization()
		a.Spec.Replicas = MarshalledReplicas(defaultReplicas)
	} else if replicas, err := GetScalingReplicas(a.Spec.Replicas); err == nil {
		if replicas.Min > replicas.Max {
			replicas.Max = replicas.Min
			a.Spec.Replicas = MarshalledReplicas(replicas)
		}
	}
}

func (a *Application) FillDefaultsStatus() {
	var msg string

	if a.Status.Summary.Status == "" {
		msg = "Default Application status, it has not initialized yet"
	} else {
		msg = "Application is trying to reconcile"
	}

	a.Status.Summary = Status{
		Status:    PENDING,
		Message:   msg,
		TimeStamp: time.Now().String(),
	}

	if a.Status.SubResources == nil {
		a.Status.SubResources = make(map[string]Status)
	}

	if len(a.Status.Conditions) == 0 {
		a.Status.Conditions = make([]metav1.Condition, 0)
	}
}

func (a *Application) GetStatus() *SkiperatorStatus {
	return &a.Status
}

func (a *Application) SetStatus(status SkiperatorStatus) {
	a.Status = status
}

// TODO clean up labels
func (a *Application) GetDefaultLabels() map[string]string {
	return map[string]string{
		"app.kubernetes.io/name":                  a.Name,
		"app.kubernetes.io/managed-by":            "skiperator",
		"skiperator.kartverket.no/controller":     "application",
		"application.skiperator.no/app":           a.Name,
		"application.skiperator.no/app-name":      a.Name,
		"application.skiperator.no/app-namespace": a.Namespace,
	}
}

func (a *Application) GetCommonSpec() *CommonSpec {
	return &CommonSpec{
		GCP:           a.Spec.GCP,
		AccessPolicy:  a.Spec.AccessPolicy,
		IstioSettings: &a.Spec.IstioSettings.IstioSettingsBase,
		Image:         a.Spec.Image,
	}
}

func (s *ApplicationSpec) Hosts() (common.HostCollection, error) {
	hosts := common.NewCollection()

	var errorsFound []error
	for _, ingress := range s.Ingresses {
		err := hosts.Add(ingress)
		if err != nil {
			errorsFound = append(errorsFound, err)
			continue
		}
	}

	return hosts, errors.Join(errorsFound...)
}

func (s *ApplicationSpec) IsRequestAuthEnabled() bool {
	return (s.IDPorten != nil && s.IDPorten.IsRequestAuthEnabled()) || (s.Maskinporten != nil && s.Maskinporten.IsRequestAuthEnabled())
}

type MultiErr interface {
	Unwrap() []error
}
