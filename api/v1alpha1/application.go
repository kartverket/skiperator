package v1alpha1

import (
	"strings"
	"time"

	"golang.org/x/exp/constraints"
	"golang.org/x/exp/slices"
	corev1 "k8s.io/api/core/v1"
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
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.application.status`
type Application struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ApplicationSpec `json:"spec,omitempty"`

	Status ApplicationStatus `json:"status,omitempty"`
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

	// Any external hostnames that route to this application. Using a skip.statkart.no-address
	// will make the application reachable for kartverket-clients (internal), other addresses
	// make the app reachable on the internet. Note that other addresses than skip.statkart.no
	// (also known as pretty hostnames) requires additional DNS setup.
	// The below hostnames will also have TLS certificates issued and be reachable on both
	// HTTP and HTTPS.
	//
	// Ingresses must be lower case, contain no spaces, be a non-empty string, and have a hostname/domain separated by a period
	//
	//+kubebuilder:validation:Optional
	Ingresses []string `json:"ingresses,omitempty"`

	// Configuration used to automatically scale the deployment based on load. If Replicas is set you must set Replicas.Min.
	//
	//+kubebuilder:validation:Optional
	Replicas Replicas `json:"replicas,omitempty"`

	// An optional priority. Supported values are 'low', 'medium' and 'high'.
	// The default value is 'medium'.
	//
	// Most workloads should not have to specify this field. If you think you
	// do, please consult with SKIP beforehand.
	//
	//+kubebuilder:validation:Enum=low;medium;high
	//+kubebuilder:default=medium
	Priority string `json:"priority,omitempty"`

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
	Resources ResourceRequirements `json:"resources,omitempty"`

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

	// Controls whether the application will automatically redirect all HTTP calls to HTTPS via the istio VirtualService.
	// This redirect does not happen on the route /.well-known/acme-challenge/, as the ACME challenge can only be done on port 80.
	//
	//+kubebuilder:validation:Optional
	//+kubebuilder:default:=true
	RedirectToHTTPS *bool `json:"redirectToHTTPS,omitempty"`


	// The root AccessPolicy for managing zero trust access to your Application. See AccessPolicy for more information.
	//
	//+kubebuilder:validation:Optional
	AccessPolicy AccessPolicy `json:"accessPolicy,omitempty"`

	// For authentication with GCP, to use services like Secret Manager and/or Pub/Sub we need
	// to set the GCP Service Account Pods should identify as. To allow this, we need the IAM role iam.workloadIdentityUser set on a GCP
	// service account and bind this to the Pod's Kubernetes SA.
	//
	// Documentation on how this is done can be found here (Closed Wiki):
	// https://kartverket.atlassian.net/wiki/spaces/SKIPDOK/pages/422346824/Autentisering+mot+GCP+som+Kubernetes+SA
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
}

// ResourceRequirements
//
// A simplified version of the Kubernetes native ResourceRequirement field, in which only Limits and Requests are present.
// For the units used for resources, see https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/#resource-units-in-kubernetes
//
// +kubebuilder:object:generate=true
type ResourceRequirements struct {

	// Limits set the maximum the app is allowed to use. Exceeding this limit will
	// make kubernetes kill the app and restart it.
	//
	// Limits can be set on the CPU and memory, but it is not recommended to put a limit on CPU, see: https://home.robusta.dev/blog/stop-using-cpu-limits
	//
	//+kubebuilder:validation:Optional
	Limits corev1.ResourceList `json:"limits,omitempty"`

	// Requests set the initial allocation that is done for the app and will
	// thus be available to the app on startup. More is allocated on demand
	// until the limit is reached.
	//
	// Requests can be set on the CPU and memory.
	//
	//+kubebuilder:validation:Optional
	Requests corev1.ResourceList `json:"requests,omitempty"`
}

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
	//+kubebuilder:validation:Optional
	TargetCpuUtilization uint `json:"targetCpuUtilization,omitempty"`
}

// Strategy
//
// Object representing a Kubernetes deployment strategy. Currently only contains a Type object,
// could probably be omitted in favour of directly using the Type.
type Strategy struct {
	// Valid values are: RollingUpdate, Recreate. Default is RollingUpdate
	//
	//+kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum=RollingUpdate;Recreate
	// +kubebuilder:default=RollingUpdate
	Type string `json:"type"`
}

type EnvFrom struct {
	// Name of Kubernetes ConfigMap in which the deployment should mount environment variables from. Must be in the same namespace as the Application
	//
	//+kubebuilder:validation:Optional
	ConfigMap string `json:"configMap,omitempty"`

	// Name of Kubernetes Secret in which the deployment should mount environment variables from. Must be in the same namespace as the Application
	//
	//+kubebuilder:validation:Optional
	Secret string `json:"secret,omitempty"`
}

// FilesFrom
//
// Struct representing information needed to mount a Kubernetes resource as a file to a Pod's directory.
// One of ConfigMap, Secret, EmptyDir or PersistentVolumeClaim must be present. 
type FilesFrom struct {
	// The path to mount the file in the Pods directory. Required.
	//
	//+kubebuilder:validation:Required
	MountPath string `json:"mountPath"`

	//+kubebuilder:validation:Optional
	ConfigMap string `json:"configMap,omitempty"`
	//+kubebuilder:validation:Optional
	Secret string `json:"secret,omitempty"`
	//+kubebuilder:validation:Optional
	EmptyDir string `json:"emptyDir,omitempty"`
	//+kubebuilder:validation:Optional
	PersistentVolumeClaim string `json:"persistentVolumeClaim,omitempty"`
}

// Probe
//
// Type configuration for all types of Kubernetes probes.
type Probe struct {
	// Number of the port to access on the container
	//
	//+kubebuilder:validation:Required
	Port uint16 `json:"port"`

	// The path to access on the HTTP server
	//
	//+kubebuilder:validation:Required
	Path string `json:"path"`

	// Delay sending the first probe by X seconds. Can be useful for applications that
	// are slow to start.
	//
	//+kubebuilder:validation:Optional
	InitialDelay uint `json:"initialDelay,omitempty"`

	// Number of seconds after which the probe times out. Defaults to 1 second.
	// Minimum value is 1
	//
	//+kubebuilder:validation:Optional
	Timeout uint `json:"timeout,omitempty"`

	// Minimum consecutive failures for the probe to be considered failed after
	// having succeeded. Defaults to 3. Minimum value is 1
	//
	//+kubebuilder:validation:Optional
	FailureThreshold uint `json:"failureThreshold,omitempty"`
}

// AccessPolicy
//
// Zero trust dictates that only applications with a reason for being able
// to access another resource should be able to reach it. This is set up by
// default by denying all ingress and egress traffic from the Pods in the
// Deployment. The AccessPolicy field is an allowlist of other applications and hostnames
// that are allowed to talk with this Application and which resources this app can talk to
//
// +kubebuilder:object:generate=true
type AccessPolicy struct {
	// Inbound specifies the ingress rules. Which apps on the cluster can talk to this app?
	//
	//+kubebuilder:validation:Optional
	Inbound InboundPolicy `json:"inbound,omitempty"`

	// Outbound specifies egress rules. Which apps on the cluster and the
	// internet is the Application allowed to send requests to?
	//
	//+kubebuilder:validation:Optional
	Outbound OutboundPolicy `json:"outbound,omitempty"`
}

// InboundPolicy
//
//+kubebuilder:object:generate=true
type InboundPolicy struct {
	// The rules list specifies a list of applications. When no namespace is
	// specified it refers to an app in the current namespace. For apps in
	// other namespaces namespace is required
	//
	//+kubebuilder:validation:Optional
	Rules []InternalRule `json:"rules"`
}

// OutboundPolicy
//
// The rules list specifies a list of applications that are reachable on the cluster.
// Note that the application you're trying to reach also must specify that they accept communication
// from this app in their ingress rules.
//
//+kubebuilder:object:generate=true
type OutboundPolicy struct {
	// Rules apply the same in-cluster rules as InboundPolicy
	//
	//+kubebuilder:validation:Optional
	Rules []InternalRule `json:"rules,omitempty"`

	// External specifies which applications on the internet the application
	// can reach. Only host is required unless it is on another port than HTTPS port 443.
	// If other ports or protocols are required then `ports` must be specified as well
	//
	//+kubebuilder:validation:Optional
	External []ExternalRule `json:"external,omitempty"`
}

// InternalRule
//
// The rules list specifies a list of applications. When no namespace is
// specified it refers to an app in the current namespace. For apps in
// other namespaces namespace is required
//
//+kubebuilder:validation:Optional
type InternalRule struct {
	// The namespace in which the Application you are allowing traffic to/from resides. If unset, uses namespace of Application.
	//
	//+kubebuilder:validation:Optional
	Namespace string `json:"namespace,omitempty"`

	// The name of the Application you are allowing traffic to/from.
	//
	//+kubebuilder:validation:Required
	Application string `json:"application"`
}

// ExternalRule
//
// Describes a rule for allowing your Application to route traffic to external applications and hosts.
//
// +kubebuilder:object:generate=true
type ExternalRule struct {
	// The allowed hostname. Note that this does not include subdomains.
	//
	//+kubebuilder:validation:Required

	Host string `json:"host"`
	// Non-HTTP requests (i.e. using the TCP protocol) need to use IP in addition to hostname
	// Only required for TCP requests.
	//
	// Note: Hostname must always be defined even if IP is set statically
	//
	//+kubebuilder:validation:Optional
	Ip string `json:"ip,omitempty"`

	// The ports to allow for the above hostname. When not specified HTTP and
	// HTTPS on port 80 and 443 respectively are put into the allowlist
	//
	//+kubebuilder:validation:Optional
	Ports []ExternalPort `json:"ports,omitempty"`
}

// ExternalPort
//
// A custom port describing an external host
type ExternalPort struct {
	// Name is required and is an arbitrary name. Must be unique within all ExternalRule ports.
	//
	//+kubebuilder:validation:Required
	Name string `json:"name"`

	// The port number of the external host
	//
	//+kubebuilder:validation:Required
	Port int `json:"port"`

	// The protocol to use for communication with the host. Only HTTP, HTTPS and TCP are supported.
	//
	//+kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=HTTP;HTTPS;TCP
	Protocol string `json:"protocol"`
}

type InternalPort struct {
	//+kubebuilder:validation:Required
	Name string `json:"name"`
	//+kubebuilder:validation:Required
	Port int32 `json:"port"`
	//+kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=TCP;UDP;SCTP
	// +kubebuilder:default:TCP
	Protocol corev1.Protocol `json:"protocol"`
}

// GCP
//
// Configuration for interacting with Google Cloud Platform
type GCP struct {
	// Configuration for authenticating a Pod with Google Cloud Platform
	//
	//+kubebuilder:validation:Required
	Auth Auth `json:"auth"`
}

// Auth
//
// Configuration for authenticating a Pod with Google Cloud Platform
type Auth struct {
	// Name of the service account in which you are trying to authenticate your pod with
	// Generally takes the form of some-name@some-project-id.iam.gserviceaccount.com
	//
	//+kubebuilder:validation:Required
	ServiceAccount string `json:"serviceAccount"`
}

// ApplicationStatus
//
// A status field shown on the Application resource which contains information regarding all controllers present on the Application.
// Will for example show errors on the Deployment field when something went wrong when attempting to create a Deployment.
//
// +kubebuilder:object:generate=true
type ApplicationStatus struct {
	ApplicationStatus Status            `json:"application"`
	ControllersStatus map[string]Status `json:"controllers"`
}

// Status
//
//+kubebuilder:object:generate=true
type Status struct {
	// +kubebuilder:default="Synced"
	Status StatusNames `json:"status"`
	// +kubebuilder:default="hello"
	Message string `json:"message"`
	// +kubebuilder:default="hello"
	TimeStamp string `json:"timestamp"`
}

type StatusNames string

const (
	SYNCED      StatusNames = "Synced"
	PROGRESSING StatusNames = "Progressing"
	ERROR       StatusNames = "Error"
	PENDING     StatusNames = "Pending"
)

func (a *Application) FillDefaultsSpec() {
	a.Spec.Replicas.Min = max(1, a.Spec.Replicas.Min)
	a.Spec.Replicas.Max = max(a.Spec.Replicas.Min, a.Spec.Replicas.Max)

	if a.Spec.Replicas.TargetCpuUtilization == 0 {
		a.Spec.Replicas.TargetCpuUtilization = 80
	}

	if a.Spec.Strategy.Type == "" {
		a.Spec.Strategy.Type = "RollingUpdate"
	}
}

func (a *Application) FillDefaultsStatus() {
	if a.Status.ApplicationStatus.Status == "" {
		a.Status.ApplicationStatus = Status{
			Status:    PENDING,
			Message:   "Default application status, application has not initialized yet",
			TimeStamp: time.Now().String(),
		}
	}

	if a.Status.ControllersStatus == nil {
		a.Status.ControllersStatus = make(map[string]Status)
	}
}

func (a *Application) UpdateApplicationStatus() {
	newApplicationStatus := a.CalculateApplicationStatus()
	if newApplicationStatus.Status == a.Status.ApplicationStatus.Status {
		return
	}

	a.Status.ApplicationStatus = newApplicationStatus
}

func (a *Application) UpdateControllerStatus(controllerName string, message string, status StatusNames) {
	if a.Status.ControllersStatus[controllerName].Status == status {
		return
	}

	newStatus := Status{
		Status:    status,
		Message:   message,
		TimeStamp: time.Now().String(),
	}
	a.Status.ControllersStatus[controllerName] = newStatus

	a.UpdateApplicationStatus()

}

func (a *Application) ShouldUpdateApplicationStatus(newStatus Status) bool {
	shouldUpdate := newStatus.Status != a.Status.ApplicationStatus.Status

	return shouldUpdate
}

func (a *Application) CalculateApplicationStatus() Status {
	returnStatus := Status{
		Status:    ERROR,
		Message:   "CALCULATION DEFAULT, YOU SHOULD NOT SEE THIS MESSAGE. PLEASE LET SKIP KNOW IF THIS MESSAGE IS VISIBLE",
		TimeStamp: time.Now().String(),
	}
	statusList := []string{}
	for _, s := range a.Status.ControllersStatus {
		statusList = append(statusList, string(s.Status))
	}

	if slices.IndexFunc(statusList, func(s string) bool { return s == string(ERROR) }) != -1 {
		returnStatus.Status = ERROR
		returnStatus.Message = "One of the controllers is in a failed state"
		return returnStatus
	}

	if slices.IndexFunc(statusList, func(s string) bool { return s == string(PROGRESSING) }) != -1 {
		returnStatus.Status = PROGRESSING
		returnStatus.Message = "One of the controllers is progressing"
		return returnStatus
	}

	if allSameStatus(statusList) {
		returnStatus.Status = StatusNames(statusList[0])
		if returnStatus.Status == SYNCED {
			returnStatus.Message = "All controllers synced"
		} else if returnStatus.Status == PENDING {
			returnStatus.Message = "All controllers pending"
		}
		return returnStatus
	}

	return returnStatus
}

func allSameStatus(a []string) bool {
	for _, v := range a {
		if v != a[0] {
			return false
		}
	}
	return true
}

func max[T constraints.Ordered](a, b T) T {
	if a > b {
		return a
	} else {
		return b
	}
}

type ControllerResources string

const (
	DEPLOYMENT              ControllerResources = "Deployment"
	SERVICE                 ControllerResources = "Service"
	SERVICEACCOUNT          ControllerResources = "ServiceAccount"
	CONFIGMAP               ControllerResources = "ConfigMap"
	NETWORKPOLICY           ControllerResources = "NetworkPolicy"
	GATEWAY                 ControllerResources = "Gateway"
	SERVICEENTRY            ControllerResources = "ServiceEntry"
	VIRTUALSERVICE          ControllerResources = "VirtualService"
	PEERAUTHENTICATION      ControllerResources = "PeerAuthentication"
	HORIZONTALPODAUTOSCALER ControllerResources = "HorizontalPodAutoscaler"
	CERTIFICATE             ControllerResources = "Certificate"
	AUTHORIZATIONPOLICY     ControllerResources = "AuthorizationPolicy"
)

func (a *Application) GroupKindFromControllerResource(controllerResource string) (metav1.GroupKind, bool) {
	switch strings.ToLower(controllerResource) {
	case "deployment":
		return metav1.GroupKind{
			Group: "apps",
			Kind:  string(DEPLOYMENT),
		}, true
	case "service":
		return metav1.GroupKind{
			Group: "",
			Kind:  string(SERVICE),
		}, true
	case "serviceaccount":
		return metav1.GroupKind{
			Group: "",
			Kind:  string(SERVICEACCOUNT),
		}, true
	case "configmaps":
		return metav1.GroupKind{
			Group: "",
			Kind:  string(CONFIGMAP),
		}, true
	case "networkpolicy":
		return metav1.GroupKind{
			Group: "networking.k8s.io",
			Kind:  string(NETWORKPOLICY),
		}, true
	case "gateway":
		return metav1.GroupKind{
			Group: "networking.istio.io",
			Kind:  string(GATEWAY),
		}, true
	case "serviceentry":
		return metav1.GroupKind{
			Group: "networking.istio.io",
			Kind:  string(SERVICEENTRY),
		}, true
	case "virtualservice":
		return metav1.GroupKind{
			Group: "networking.istio.io",
			Kind:  string(VIRTUALSERVICE),
		}, true
	case "peerauthentication":
		return metav1.GroupKind{
			Group: "security.istio.io",
			Kind:  string(PEERAUTHENTICATION),
		}, true
	case "horizontalpodautoscaler":
		return metav1.GroupKind{
			Group: "autoscaling",
			Kind:  string(HORIZONTALPODAUTOSCALER),
		}, true
	case "certificate":
		return metav1.GroupKind{
			Group: "cert-manager.io",
			Kind:  string(CERTIFICATE),
		}, true
	case "authorizationpolicy":
		return metav1.GroupKind{
			Group: "security.istio.io",
			Kind:  string(AUTHORIZATIONPOLICY),
		}, true
	default:
		return metav1.GroupKind{}, false
	}
}
