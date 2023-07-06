package v1alpha1

import (
	"golang.org/x/exp/constraints"
	"k8s.io/apimachinery/pkg/util/intstr"
	"strings"
	"time"

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
	//+kubebuilder:validation:Required
	Image string `json:"image"`
	//+kubebuilder:validation:Enum=low;medium;high
	//+kubebuilder:default=medium
	Priority string `json:"priority,omitempty"`
	//+kubebuilder:validation:Optional
	Command []string `json:"command,omitempty"`

	//+kubebuilder:validation:Optional
	Resources ResourceRequirements `json:"resources,omitempty"`
	//+kubebuilder:validation:Optional
	Replicas *Replicas `json:"replicas,omitempty"`
	//+kubebuilder:validation:Optional
	Strategy Strategy `json:"strategy,omitempty"`

	//+kubebuilder:validation:Optional
	Env []corev1.EnvVar `json:"env,omitempty"`
	//+kubebuilder:validation:Optional
	EnvFrom []EnvFrom `json:"envFrom,omitempty"`
	//+kubebuilder:validation:Optional
	FilesFrom []FilesFrom `json:"filesFrom,omitempty"`

	//+kubebuilder:validation:Required
	Port int `json:"port"`
	//+kubebuilder:validation:Optional
	AdditionalPorts []InternalPort `json:"additionalPorts,omitempty"`
	//+kubebuilder:validation:Optional
	Prometheus *PrometheusConfig `json:"prometheus,omitempty"`
	//+kubebuilder:validation:Optional
	Liveness *Probe `json:"liveness,omitempty"`
	//+kubebuilder:validation:Optional
	Readiness *Probe `json:"readiness,omitempty"`
	//+kubebuilder:validation:Optional
	Startup *Probe `json:"startup,omitempty"`

	// Ingresses must be lower case, contain no spaces, be a non-empty string, and have a hostname/domain separated by a period
	//
	//+kubebuilder:validation:Optional
	Ingresses []string `json:"ingresses,omitempty"`

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

	//+kubebuilder:validation:Optional
	AccessPolicy AccessPolicy `json:"accessPolicy,omitempty"`

	//+kubebuilder:validation:Optional
	GCP *GCP `json:"gcp,omitempty"`

	//+kubebuilder:validation:Optional
	Labels map[string]string `json:"labels,omitempty"`

	//+kubebuilder:validation:Optional
	ResourceLabels map[string]map[string]string `json:"resourceLabels,omitempty"`

	//+kubebuilder:validation:Optional
	AuthorizationSettings *AuthorizationSettings `json:"authorizationSettings,omitempty"`
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

// ResourceRequirements
// +kubebuilder:object:generate=true
type ResourceRequirements struct {
	// TODO
	// Remember to reassess whether or not Claims work properly with kubebuilder when we upgrade to Kubernetes 1.26

	//+kubebuilder:validation:Optional
	Limits corev1.ResourceList `json:"limits,omitempty"`

	//+kubebuilder:validation:Optional
	Requests corev1.ResourceList `json:"requests,omitempty"`
}

// +kubebuilder:object:generate=true
type Replicas struct {
	//+kubebuilder:validation:Required
	Min uint `json:"min"`
	//+kubebuilder:validation:Optional
	Max uint `json:"max,omitempty"`

	//+kubebuilder:default:=80
	//+kubebuilder:validation:Optional
	TargetCpuUtilization uint `json:"targetCpuUtilization,omitempty"`
}

// +kubebuilder:object:generate=true
type Strategy struct {
	// +kubebuilder:default:=RollingUpdate
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum=RollingUpdate;Recreate
	Type string `json:"type,omitempty"`
}

type EnvFrom struct {
	//+kubebuilder:validation:Optional
	ConfigMap string `json:"configMap,omitempty"`
	//+kubebuilder:validation:Optional
	Secret string `json:"secret,omitempty"`
}

type FilesFrom struct {
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

// PrometheusConfig contains configuration settings instructing how the app should be scraped.
// +kubebuilder:object:generate=true
type PrometheusConfig struct {
	// The port number or name where metrics are exposed (at the Pod level).
	//+kubebuilder:validation:Required
	Port intstr.IntOrString `json:"port"`
	// The HTTP path where Prometheus compatible metrics exists
	//+kubebuilder:default:=/metrics
	//+kubebuilder:validation:Optional
	Path string `json:"path,omitempty"`
}

// +kubebuilder:object:generate=true
type Probe struct {
	//+kubebuilder:default=0
	//+kubebuilder:validation:Optional
	InitialDelay int32 `json:"initialDelay,omitempty"`
	//+kubebuilder:default=1
	//+kubebuilder:validation:Optional
	Timeout int32 `json:"timeout,omitempty"`
	//+kubebuilder:default=10
	//+kubebuilder:validation:Optional
	Period int32 `json:"period,omitempty"`
	//+kubebuilder:default=1
	//+kubebuilder:validation:Optional
	SuccessThreshold int32 `json:"successThreshold,omitempty"`
	//+kubebuilder:default=3
	//+kubebuilder:validation:Optional
	FailureThreshold int32 `json:"failureThreshold,omitempty"`
	//+kubebuilder:validation:Required
	Port uint16 `json:"port"`
	//+kubebuilder:validation:Required
	Path string `json:"path"`
}

// +kubebuilder:object:generate=true
type AccessPolicy struct {
	//+kubebuilder:validation:Optional
	Inbound InboundPolicy `json:"inbound,omitempty"`
	//+kubebuilder:validation:Optional
	Outbound OutboundPolicy `json:"outbound,omitempty"`
}

// +kubebuilder:object:generate=true
type InboundPolicy struct {
	//+kubebuilder:validation:Optional
	Rules []InternalRule `json:"rules"`
}

// +kubebuilder:object:generate=true
type OutboundPolicy struct {
	//+kubebuilder:validation:Optional
	Rules []InternalRule `json:"rules,omitempty"`
	//+kubebuilder:validation:Optional
	External []ExternalRule `json:"external,omitempty"`
}

type InternalRule struct {
	//+kubebuilder:validation:Optional
	Namespace string `json:"namespace,omitempty"`
	//+kubebuilder:validation:Required
	Application string `json:"application"`
}

// +kubebuilder:object:generate=true
type ExternalRule struct {
	//+kubebuilder:validation:Required
	Host string `json:"host"`
	//+kubebuilder:validation:Optional
	Ip string `json:"ip,omitempty"`
	//+kubebuilder:validation:Optional
	Ports []ExternalPort `json:"ports,omitempty"`
}

type ExternalPort struct {
	//+kubebuilder:validation:Required
	Name string `json:"name"`
	//+kubebuilder:validation:Required
	Port int `json:"port"`
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

type GCP struct {
	//+kubebuilder:validation:Required
	Auth Auth `json:"auth"`
}

type Auth struct {
	//+kubebuilder:validation:Required
	ServiceAccount string `json:"serviceAccount"`
}

// +kubebuilder:object:generate=true
type ApplicationStatus struct {
	ApplicationStatus Status            `json:"application"`
	ControllersStatus map[string]Status `json:"controllers"`
}

// +kubebuilder:object:generate=true
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
	if a.Spec.Replicas == nil {
		a.Spec.Replicas = &Replicas{
			Min:                  2,
			Max:                  5,
			TargetCpuUtilization: 80,
		}
	} else if a.Spec.Replicas.Min == 0 && a.Spec.Replicas.Max == 0 {
	} else {
		a.Spec.Replicas.Min = max(1, a.Spec.Replicas.Min)
		a.Spec.Replicas.Max = max(a.Spec.Replicas.Min, a.Spec.Replicas.Max)
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
