package config_patch

type OAuthClusterConfigPatchValue struct {
	Name            string          `json:"name"`
	DnsLookupFamily string          `json:"dns_lookup_family"`
	Type            string          `json:"type"`
	ConnectTimeout  string          `json:"connect_timeout"`
	LbPolicy        string          `json:"lb_policy"`
	TransportSocket TransportSocket `json:"transport_socket"`
	LoadAssignment  LoadAssignment  `json:"load_assignment"`
}

type TransportSocket struct {
	Name        string      `json:"name"`
	TypedConfig TypedConfig `json:"typed_config"`
}

type TypedConfig struct {
	Type   string  `json:""@type""`
	Sni    *string `json:"sni,omitempty"`
	Config *Config `json:"config,omitempty"`
}

type LoadAssignment struct {
	ClusterName string                   `json:"cluster_name"`
	Endpoints   []LoadAssignmentEndpoint `json:"endpoints"`
}

type LoadAssignmentEndpoint struct {
	LbEndpoints []LbEndpoint `json:"lb_endpoints"`
}

type LbEndpoint struct {
	Endpoint Endpoint `json:"endpoint"`
}

type Endpoint struct {
	Address Address `json:"address"`
}

type Address struct {
	SocketAddress SocketAddress `json:"socket_address"`
}

type SocketAddress struct {
	Address   string `json:"address"`
	PortValue int    `json:"port_value"`
}

func GetOAuthClusterConfigPatchValue(identityProviderHost string) OAuthClusterConfigPatchValue {
	return OAuthClusterConfigPatchValue{
		Name:            "oauth",
		DnsLookupFamily: "AUTO",
		Type:            "LOGICAL_DNS",
		ConnectTimeout:  "10s",
		LbPolicy:        "ROUND_ROBIN",
		TransportSocket: TransportSocket{
			Name: "envoy.transport_sockets.tls",
			TypedConfig: TypedConfig{
				Type: "type.googleapis.com/envoy.extensions.transport_sockets.tls.v3.UpstreamTlsContext",
				Sni:  &identityProviderHost,
			},
		},
		LoadAssignment: LoadAssignment{
			ClusterName: "oauth",
			Endpoints: []LoadAssignmentEndpoint{
				{
					LbEndpoints: []LbEndpoint{
						{
							Endpoint: Endpoint{
								Address: Address{
									SocketAddress: SocketAddress{
										Address:   identityProviderHost,
										PortValue: 443,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}
