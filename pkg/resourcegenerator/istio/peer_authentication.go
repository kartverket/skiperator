package istio

import (
	"github.com/kartverket/skiperator/pkg/util"
	securityv1beta1api "istio.io/api/security/v1beta1"
	typev1beta1 "istio.io/api/type/v1beta1"
)

func GetPeerAuthentication(ownerName string) securityv1beta1api.PeerAuthentication {
	return securityv1beta1api.PeerAuthentication{
		Selector: &typev1beta1.WorkloadSelector{
			MatchLabels: util.GetPodAppSelector(ownerName),
		},
		Mtls: &securityv1beta1api.PeerAuthentication_MutualTLS{
			Mode: securityv1beta1api.PeerAuthentication_MutualTLS_STRICT,
		},
	}
}
