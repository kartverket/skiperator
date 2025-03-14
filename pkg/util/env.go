package util

type RegistryCredentials struct {
	Registry string `env:"REGISTRY"`
	Token    string `env:"TOKEN"`
}
type EnvVars struct {
	RegistryCredentialsList []RegistryCredentials `envPrefix:"IMAGE_PULL"`
}
