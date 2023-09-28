package podtypes

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
