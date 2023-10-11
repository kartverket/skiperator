package v1alpha1

import (
	"time"

	"encoding/json"
	"github.com/kartverket/skiperator/api/v1alpha1/podtypes"
	"golang.org/x/exp/constraints"
	"k8s.io/apimachinery/pkg/util/intstr"

	nais_io_v1 "github.com/nais/liberator/pkg/apis/nais.io/v1"
	"golang.org/x/exp/slices"
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
	// Ingresses must be lowercase, contain no spaces, be a non-empty string, and have a hostname/domain separated by a period
	//
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
	Resources *podtypes.ResourceRequirements `json:"resources,omitempty"`

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
	EnvFrom []podtypes.EnvFrom `json:"envFrom,omitempty"`

	// Mounting volumes into the Deployment are done using the FilesFrom argument
	//
	// FilesFrom supports ConfigMaps, Secrets and PVCs. The Application resource
	// assumes these have already been created by you, and will fail if this is not the case.
	//
	// For mounting environment variables see EnvFrom.
	//
	//+kubebuilder:validation:Optional
	FilesFrom []podtypes.FilesFrom `json:"filesFrom,omitempty"`

	// An optional list of extra port to expose on a pod level basis,
	// for example so Instana or other APM tools can reach it
	//
	//+kubebuilder:validation:Optional
	AdditionalPorts []podtypes.InternalPort `json:"additionalPorts,omitempty"`
	// Liveness probes define a resource that returns 200 OK when the app is running
	// as intended. Returning a non-200 code will make kubernetes restart the app.
	// Liveness is optional, but when provided, path and port are required
	//
	// See Probe for structure definition.
	//
	//+kubebuilder:validation:Optional
	Liveness *podtypes.Probe `json:"liveness,omitempty"`

	// Readiness probes define a resource that returns 200 OK when the app is running
	// as intended. Kubernetes will wait until the resource returns 200 OK before
	// marking the pod as Running and progressing with the deployment strategy.
	// Readiness is optional, but when provided, path and port are required
	//
	//+kubebuilder:validation:Optional
	Readiness *podtypes.Probe `json:"readiness,omitempty"`

	// Kubernetes uses startup probes to know when a container application has started.
	// If such a probe is configured, it disables liveness and readiness checks until it
	// succeeds, making sure those probes don't interfere with the application startup.
	// This can be used to adopt liveness checks on slow starting containers, avoiding them
	// getting killed by Kubernetes before they are up and running.
	// Startup is optional, but when provided, path and port are required
	//
	//+kubebuilder:validation:Optional
	Startup *podtypes.Probe `json:"startup,omitempty"`

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
	AccessPolicy *podtypes.AccessPolicy `json:"accessPolicy,omitempty"`

	// For authentication with GCP, to use services like Secret Manager and/or Pub/Sub we need
	// to set the GCP Service Account Pods should identify as. To allow this, we need the IAM role iam.workloadIdentityUser set on a GCP
	// service account and bind this to the Pod's Kubernetes SA.
	//
	// Documentation on how this is done can be found here (Closed Wiki):
	// https://kartverket.atlassian.net/wiki/spaces/SKIPDOK/pages/422346824/Autentisering+mot+GCP+som+Kubernetes+SA
	//
	//+kubebuilder:validation:Optional
	GCP *podtypes.GCP `json:"gcp,omitempty"`

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

// Based off NAIS' IDPorten specification as seen here:
// https://github.com/nais/liberator/blob/c9da4cf48a52c9594afc8a4325ff49bbd359d9d2/pkg/apis/nais.io/v1/naiserator_types.go#L93C10-L93C10
//
// +kubebuilder:object:generate=true
type IDPorten struct {
	// The name of the Client as shown in Digitaliseringsdirektoratet's Samarbeidsportal
	// Meant to be a human-readable name for separating clients in the portal
	ClientName *string `json:"clientName,omitempty"`

	// Whether to enable provisioning of an ID-porten client.
	// If enabled, an ID-porten client be provisioned.
	Enabled bool `json:"enabled"`

	// AccessTokenLifetime is the lifetime in seconds for any issued access token from ID-porten.
	//
	// If unspecified, defaults to `3600` seconds (1 hour).
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=3600
	AccessTokenLifetime *int `json:"accessTokenLifetime,omitempty"`

	// ClientURI is the URL shown to the user at ID-porten when displaying a 'back' button or on errors.
	ClientURI nais_io_v1.IDPortenURI `json:"clientURI,omitempty"`

	// FrontchannelLogoutPath is a valid path for your application where ID-porten sends a request to whenever the user has
	// initiated a logout elsewhere as part of a single logout (front channel logout) process.
	//
	// +kubebuilder:validation:Pattern=`^\/.*$`
	FrontchannelLogoutPath string `json:"frontchannelLogoutPath,omitempty"`

	// IntegrationType is used to make sensible choices for your client.
	// Which type of integration you choose will provide guidance on which scopes you can use with the client.
	// A client can only have one integration type.
	//
	// NB! It is not possible to change the integration type after creation.
	//
	// +kubebuilder:validation:Enum=krr;idporten;api_klient
	IntegrationType string `json:"integrationType,omitempty" nais:"immutable"`

	// PostLogoutRedirectPath is a simpler verison of PostLogoutRedirectURIs
	// that will be appended to the ingress
	//
	// +kubebuilder:validation:Pattern=`^\/.*$`
	// +kubebuilder:validation:Optional
	PostLogoutRedirectPath string `json:"postLogoutRedirectPath,omitempty"`

	// PostLogoutRedirectURIs are valid URIs that ID-porten will allow redirecting the end-user to after a single logout
	// has been initiated and performed by the application.
	PostLogoutRedirectURIs *[]nais_io_v1.IDPortenURI `json:"postLogoutRedirectURIs,omitempty"`

	// RedirectPath is a valid path that ID-porten redirects back to after a successful authorization request.
	//
	// +kubebuilder:validation:Pattern=`^\/.*$`
	// +kubebuilder:validation:Optional
	RedirectPath string `json:"redirectPath,omitempty"`

	// Register different oauth2 Scopes on your client.
	// You will not be able to add a scope to your client that conflicts with the client's IntegrationType.
	// For example, you can not add a scope that is limited to the IntegrationType `krr` of IntegrationType `idporten`, and vice versa.
	//
	// Default for IntegrationType `krr` = ("krr:global/kontaktinformasjon.read", "krr:global/digitalpost.read")
	// Default for IntegrationType `idporten` = ("openid", "profile")
	// IntegrationType `api_klient` have no Default, checkout Digdir documentation.
	Scopes []string `json:"scopes,omitempty"`

	// SessionLifetime is the maximum lifetime in seconds for any given user's session in your application.
	// The timeout starts whenever the user is redirected from the `authorization_endpoint` at ID-porten.
	//
	// If unspecified, defaults to `7200` seconds (2 hours).
	// Note: Attempting to refresh the user's `access_token` beyond this timeout will yield an error.
	//
	// +kubebuilder:validation:Minimum=3600
	// +kubebuilder:validation:Maximum=7200
	SessionLifetime *int `json:"sessionLifetime,omitempty"`
}

// https://github.com/nais/liberator/blob/c9da4cf48a52c9594afc8a4325ff49bbd359d9d2/pkg/apis/nais.io/v1/naiserator_types.go#L376
// +kubebuilder:object:generate=true
type Maskinporten struct {
	// The name of the Client as shown in Digitaliseringsdirektoratet's Samarbeidsportal
	// Meant to be a human-readable name for separating clients in the portal
	ClientName *string `json:"clientName,omitempty"`

	// If enabled, provisions and configures a Maskinporten client with consumed scopes and/or Exposed scopes with DigDir.
	Enabled bool `json:"enabled"`

	// Schema to configure Maskinporten clients with consumed scopes and/or exposed scopes.
	Scopes *nais_io_v1.MaskinportenScope `json:"scopes,omitempty"`
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

// PrometheusConfig contains configuration settings instructing how the app should be scraped.
//
// +kubebuilder:object:generate=true
type PrometheusConfig struct {
	// The port number or name where metrics are exposed (at the Pod level).
	//
	//+kubebuilder:validation:Required
	Port intstr.IntOrString `json:"port"`
	// The HTTP path where Prometheus compatible metrics exists
	//
	//+kubebuilder:default:=/metrics
	//+kubebuilder:validation:Optional
	Path string `json:"path,omitempty"`
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

func NewDefaultReplicas() Replicas {
	return Replicas{
		Min:                  2,
		Max:                  5,
		TargetCpuUtilization: 80,
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
		a.Spec.Replicas = MarshalledReplicas(defaultReplicas)
	} else if replicas, err := GetScalingReplicas(a.Spec.Replicas); err == nil {
		if replicas.Min > replicas.Max {
			replicas.Max = replicas.Min
			a.Spec.Replicas = MarshalledReplicas(replicas)
		}
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
