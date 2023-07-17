package podtypes

type GCP struct {
	//+kubebuilder:validation:Required
	Auth Auth `json:"auth"`
}

type Auth struct {
	//+kubebuilder:validation:Required
	ServiceAccount string `json:"serviceAccount"`
}
