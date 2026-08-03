package serviceentry

import (
	"testing"

	"github.com/kartverket/skiperator/api/common/istiotypes"
	"github.com/kartverket/skiperator/api/common/podtypes"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/testutil"
	"github.com/stretchr/testify/assert"
	networkingv1api "istio.io/api/networking/v1"
)

func TestGetServiceEntryEndpointData(t *testing.T) {
	tests := []struct {
		name               string
		host               string
		ip                 string
		ports              []*networkingv1api.ServicePort
		expectedResolution networkingv1api.ServiceEntry_Resolution
		expectedAddresses  []string
		expectedEndpoint   string
		expectedError      string
	}{
		{
			name:               "plain host uses DNS resolution",
			host:               "www.google.com",
			expectedResolution: networkingv1api.ServiceEntry_DNS,
		},
		{
			name:               "wildcard host uses NONE resolution",
			host:               "*.google.com",
			ports:              []*networkingv1api.ServicePort{{Protocol: "HTTPS"}},
			expectedResolution: networkingv1api.ServiceEntry_NONE,
		},
		{
			name:          "wildcard TCP without static IP is rejected",
			host:          "*.google.com",
			ports:         []*networkingv1api.ServicePort{{Protocol: "TCP"}},
			expectedError: "static IP must be set for TCP port with wildcard host",
		},
		{
			name:               "static IP uses STATIC resolution",
			host:               "*.google.com",
			ip:                 "1.2.3.4",
			ports:              []*networkingv1api.ServicePort{{Protocol: "TCP"}},
			expectedResolution: networkingv1api.ServiceEntry_STATIC,
			expectedAddresses:  []string{"1.2.3.4"},
			expectedEndpoint:   "1.2.3.4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolution, addresses, endpoints, err := getServiceEntryEndpointData(tt.host, tt.ip, tt.ports)

			if tt.expectedError != "" {
				assert.ErrorContains(t, err, tt.expectedError)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedResolution, resolution)
			assert.Equal(t, tt.expectedAddresses, addresses)

			if tt.expectedEndpoint == "" {
				assert.Nil(t, endpoints)
				return
			}

			if assert.Len(t, endpoints, 1) {
				assert.Equal(t, tt.expectedEndpoint, endpoints[0].Address)
			}
		})
	}
}

func TestGenerateRejectsWildcardTCPWithoutStaticIP(t *testing.T) {
	r := testutil.GetTestMinimalAppReconciliation()
	application := r.GetSKIPObject().(*skiperatorv1alpha1.Application)
	application.Spec.IstioSettings = &istiotypes.IstioSettingsApplication{}
	application.Spec.AccessPolicy = &podtypes.AccessPolicy{
		Outbound: &podtypes.OutboundPolicy{
			External: []podtypes.ExternalRule{
				{
					Host: "*.google.com",
					Ports: []podtypes.ExternalPort{
						{Name: "tcp", Port: 1234, Protocol: "TCP"},
					},
				},
			},
		},
	}

	err := Generate(r)

	assert.ErrorContains(t, err, "static IP must be set for TCP port")
	assert.Empty(t, r.GetResources())
}
