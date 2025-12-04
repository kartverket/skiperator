package envconfig

type RegistryCredentialPair struct {
	Registry string `env:"REGISTRY"`
	Token    string `env:"TOKEN"`
}
type Vars struct {
	RegistryCredentials         []RegistryCredentialPair `envPrefix:"IMAGE_PULL"`
	ClusterCIDRExclusionEnabled bool                     `env:"CLUSTER_CIDR_EXCLUDE" envDefault:"false"`
}
