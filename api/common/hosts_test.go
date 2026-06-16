package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHostNormalizesHostname(t *testing.T) {
	host, err := NewHost("API.example.COM+custom-tls")

	require.NoError(t, err)
	assert.Equal(t, "api.example.com", host.Hostname)
	require.NotNil(t, host.CustomCertificateSecret)
	assert.Equal(t, "custom-tls", *host.CustomCertificateSecret)
}

func TestHostCollectionDeduplicatesNormalizedHostnames(t *testing.T) {
	hosts := NewCollection()

	require.NoError(t, hosts.Add("API.example.com"))
	require.NoError(t, hosts.Add("api.EXAMPLE.com"))

	assert.Equal(t, 1, hosts.Count())
	assert.Equal(t, []string{"api.example.com"}, hosts.Hostnames())
}
