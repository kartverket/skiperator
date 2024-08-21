package podtypes

// GCP
//
// Configuration for interacting with Google Cloud Platform
// +kubebuilder:object:generate=true
type GCP struct {
	// Configuration for authenticating a Pod with Google Cloud Platform
	// For authentication with GCP, to use services like Secret Manager and/or Pub/Sub we need
	// to set the GCP Service Account Pods should identify as. To allow this, we need the IAM role iam.workloadIdentityUser set on a GCP
	// service account and bind this to the Pod's Kubernetes SA.
	// Documentation on how this is done can be found here (Closed Wiki):
	// https://kartverket.atlassian.net/wiki/spaces/SKIPDOK/pages/422346824/Autentisering+mot+GCP+som+Kubernetes+SA
	//
	//+kubebuilder:validation:Optional
	Auth *Auth `json:"auth,omitempty"`

	// CloudSQL is used to deploy a CloudSQL proxy sidecar in the pod.
	// This is useful for connecting to CloudSQL databases that require Cloud SQL Auth Proxy.
	//
	//+kubebuilder:validation:Optional
	CloudSQLProxy *CloudSQLProxySettings `json:"cloudSqlProxy,omitempty"`
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

type CloudSQLProxySettings struct {
	// Connection name for the CloudSQL instance. Found in the Google Cloud Console under your CloudSQL resource.
	// The format is "projectName:region:instanceName" E.g. "skip-prod-bda1:europe-north1:my-db".
	//+kubebuilder:validation:Required
	//+kubebuilder:validation:Pattern=`^[^:]+:[^:]+:[^:]+$`
	ConnectionName string `json:"connectionName,omitempty"`

	// Service account used by cloudsql auth proxy. This service account must have the roles/cloudsql.client role.
	//+kubebuilder:validation:Required
	ServiceAccount string `json:"serviceAccount,omitempty"`

	// The IP address of the CloudSQL instance. This is used to create a serviceentry for the CloudSQL proxy.
	//+kubebuilder:validation:Required
	IP string `json:"ip,omitempty"`

	// Image version for the CloudSQL proxy sidecar.
	//+kubebuilder:validation:Optional
	//+kubebuilder:default:="2.8.0"
	Version string `json:"version"`
}
