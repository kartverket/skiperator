package resourceutils

import (
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"strings"
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

func GetImageVersion(imageVersionString string) string {
	parts := strings.Split(imageVersionString, ":")

	// Implicitly assume version "latest" if no version is specified
	if len(parts) < 2 {
		return "latest"
	}

	versionPart := parts[1]

	// Remove "@sha256" from version text if version includes a hash
	if strings.Contains(versionPart, "@") {
		versionPart = strings.Split(versionPart, "@")[0]
	}

	// Add build number to version if it is specified
	if len(parts) > 2 {
		return versionPart + "-" + parts[2]
	}
	return versionPart
}
