package resourceutils

import (
	"github.com/kartverket/skiperator/api/v1alpha1"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

func ShouldScaleToZero(jsonReplicas *apiextensionsv1.JSON) bool {
	replicas, err := skiperatorv1alpha1.GetStaticReplicas(jsonReplicas)
	if err == nil && replicas == 0 {
		return true
	}
	replicasStruct, err := skiperatorv1alpha1.GetScalingReplicas(jsonReplicas)
	if err == nil && (replicasStruct.Min == 0 || replicasStruct.Max == 0) {
		return true
	}
	return false
}

func GetIstioGatewayLabelSelector(host *v1alpha1.Host) map[string]string {
	if host.Internal {
		return map[string]string{"app": "istio-ingress-internal"}
	}
	return map[string]string{"app": "istio-ingress-external"}
}
