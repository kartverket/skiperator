package v1alpha1

import (
	"fmt"
	"github.com/chmike/domain"
	"strings"
)

const hostnameSecretSeparator = "+"

// TODO: Add a mechanism for validating that the
// hostname is covered by the CustomCertificateSecret if present
type Host struct {
	Hostname                string
	CustomCertificateSecret *string
}

func NewHost(hostname string) (*Host, error) {
	if len(hostname) == 0 {
		return nil, fmt.Errorf("hostname cannot be empty")
	}

	var h Host
	// If hostname is separated by +, the user wants to use a custom certificate
	results := strings.Split(hostname, hostnameSecretSeparator)

	switch len(results) {
	// No custom cert present
	case 1:
		h = Host{Hostname: results[0], CustomCertificateSecret: nil}
	// Custom cert present
	case 2:
		secret := results[1]
		if len(secret) == 0 {
			return nil, fmt.Errorf("%s: not valid, custom certificate secret cannot be empty", hostname)
		}

		h = Host{Hostname: results[0], CustomCertificateSecret: &secret}
	// More than one '+' characters present
	default:
		return nil, fmt.Errorf("%s: not valid, contains multiple '%s' characters", hostname, hostnameSecretSeparator)
	}

	// Verify that the hostname is an actual valid DNS name.
	if err := domain.Check(h.Hostname); err != nil {
		return nil, fmt.Errorf("%s: failed validation: %w", h.Hostname, err)
	}

	return &h, nil
}

func (h *Host) UsesCustomCert() bool {
	return h.CustomCertificateSecret != nil
}
