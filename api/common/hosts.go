package common

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/chmike/domain"
)

const hostnameSecretSeparator = "+"

// internalHostnamePattern mirrors pkg/util.internalPattern and is duplicated
// here to avoid an import cycle (api/common ↔ pkg/util). Keep in sync.
var internalHostnamePattern = regexp.MustCompile(`(?i)[^.]\.(?:skip\.statkart\.no|kartverket-intern\.cloud)$`)

// TODO: Add a mechanism for validating that the
// hostname is covered by the CustomCertificateSecret if present
type Host struct {
	Hostname                string
	CustomCertificateSecret *string
	// ForceInternal routes this host through the internal ingress gateway even
	// when the hostname does not match the known-internal domain suffixes.
	ForceInternal bool
}

// IsInternal returns true when the host should be treated as an internal
// endpoint — either because the hostname matches the known-internal domain
// pattern or because ForceInternal has been explicitly set.
func (h *Host) IsInternal() bool {
	return h.ForceInternal || internalHostnamePattern.MatchString(h.Hostname)
}


type HostCollection struct {
	hosts           map[string]*Host
	hostInsertOrder []string
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
		h = Host{Hostname: strings.ToLower(results[0]), CustomCertificateSecret: nil}
	// Custom cert present
	case 2:
		secret := results[1]
		if len(secret) == 0 {
			return nil, fmt.Errorf("%s: not valid, custom certificate secret cannot be empty", hostname)
		}

		h = Host{Hostname: strings.ToLower(results[0]), CustomCertificateSecret: &secret}
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
		hosts:           map[string]*Host{},
		hostInsertOrder: []string{},
	}
}

func (hs *HostCollection) Add(hostname string) error {
	h, err := NewHost(hostname)
	if err != nil {
		return err
	}

	existingValue, alreadyPresent := hs.hosts[h.Hostname]

	if alreadyPresent {
		if existingValue.UsesCustomCert() {
			return fmt.Errorf("host '%s' is already defined and using a custom certificate", existingValue.Hostname)
		}
	} else {
		hs.hostInsertOrder = append(hs.hostInsertOrder, h.Hostname)
	}
	hs.hosts[h.Hostname] = h
	return nil
}

func (hs *HostCollection) AllHosts() []*Host {
	hosts := make([]*Host, 0, len(hs.hosts))
	for _, hostname := range hs.hostInsertOrder {
		hosts = append(hosts, hs.hosts[hostname])
	}
	return hosts
}

func (hs *HostCollection) Hostnames() []string {
	hostnames := make([]string, len(hs.hostInsertOrder))
	copy(hostnames, hs.hostInsertOrder)
	return hostnames
}

func (hs *HostCollection) Count() int {
	return len(hs.hosts)
}
