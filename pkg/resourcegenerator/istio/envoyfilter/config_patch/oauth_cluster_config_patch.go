package config_patch

func GetOAuthClusterConfigPatchValue(identityProviderHost string) map[string]interface{} {
	return map[string]interface{}{
		"name":              "oauth",
		"dns_lookup_family": "AUTO",
		"type":              "LOGICAL_DNS",
		"connect_timeout":   "10s",
		"lb_policy":         "ROUND_ROBIN",
		"transport_socket": map[string]interface{}{
			"name": "envoy.transport_sockets.tls",
			"typed_config": map[string]interface{}{
				"@type": "type.googleapis.com/envoy.extensions.transport_sockets.tls.v3.UpstreamTlsContext",
				"sni":   identityProviderHost,
			},
		},
		"load_assignment": map[string]interface{}{
			"cluster_name": "oauth",
			"endpoints": []interface{}{
				map[string]interface{}{
					"lb_endpoints": []interface{}{
						map[string]interface{}{
							"endpoint": map[string]interface{}{
								"address": map[string]interface{}{
									"socket_address": map[string]interface{}{
										"address":    identityProviderHost,
										"port_value": 443,
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
