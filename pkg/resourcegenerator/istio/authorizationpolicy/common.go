package authorizationpolicy

import (
	"github.com/kartverket/skiperator/pkg/mesh"
	v1 "istio.io/api/security/v1"
)

const (
	DefaultDenyPath = "/actuator*"
)

func GetGeneralFromRule() []*v1.Rule_From {
	return []*v1.Rule_From{
		{
			Source: &v1.Source{
				Namespaces: []string{mesh.GatewayNamespace},
			},
		},
	}
}
