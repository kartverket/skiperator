package v1alpha1

import (
	"fmt"
	"github.com/chmike/domain"
	"regexp"
	"strings"
)

const hostnameSecretSeparator = "+"

var internalPattern = regexp.MustCompile(`[^.]\.skip\.statkart\.no|[^.]\.kartverket-intern.cloud`)

// TODO: Add a mechanism for validating that the
// hostname is covered by the CustomCertificateSecret if present
type Host struct {
	Hostname                string
	CustomCertificateSecret *string
	Internal                bool
}

type HostCollection struct {
	hosts map[string]*Host
}

func NewHost(hostname string) (*Host, error) {
	if len(hostname) == 0 {
		return nil, fmt.Errorf("hostname cannot be empty")
	}

	var h Host

	h.Internal = IsInternal(hostname)

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

func NewCollection() HostCollection {
	return HostCollection{
		hosts: map[string]*Host{},
	}
}

func (hs *HostCollection) Add(hostname string) error {
	h, err := NewHost(hostname)
	if err != nil {
		return err
	}

	existingValue, alreadyPresent := hs.hosts[h.Hostname]
	switch alreadyPresent {
	case true:
		if existingValue.UsesCustomCert() {
			return fmt.Errorf("host '%s' is already defined and using a custom certificate", existingValue.Hostname)
		}
		fallthrough
	case false:
		fallthrough
	default:
		hs.hosts[h.Hostname] = h
	}

	return nil
}

func (hs *HostCollection) AddObject(hostname string, internal bool) error {
	h, err := NewHost(hostname)
	if err != nil {
		return err
	}

	h.Internal = internal
	existingValue, alreadyPresent := hs.hosts[h.Hostname]
	switch alreadyPresent {
	case true:
		if existingValue.UsesCustomCert() {
			return fmt.Errorf("host '%s' is already defined and using a custom certificate", existingValue.Hostname)
		}
		fallthrough
	case false:
		fallthrough
	default:
		hs.hosts[h.Hostname] = h
	}

	return nil
}

func (hs *HostCollection) AllHosts() []*Host {
	hosts := make([]*Host, 0, len(hs.hosts))
	for _, host := range hs.hosts {
		hosts = append(hosts, host)
	}
	return hosts
}

func (hs *HostCollection) Hostnames() []string {
	hostnames := make([]string, 0, len(hs.hosts))
	for hostname := range hs.hosts {
		hostnames = append(hostnames, hostname)
	}
	return hostnames
}

func (hs *HostCollection) Count() int {
	return len(hs.hosts)
}

func IsInternal(hostname string) bool {
	return internalPattern.MatchString(hostname)
}
