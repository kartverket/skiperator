package envconfig

type RegistryCredentialPair struct {
	Registry string `env:"REGISTRY"`
	Token    string `env:"TOKEN"`
}
type Vars struct {
	RegistryCredentials []RegistryCredentialPair `envPrefix:"IMAGE_PULL"`
}
