package resourceutils

import (
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"regexp"
	"strings"
)

const LabelValueMaxLength int = 63

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

// MatchesRegex checks if a string matches a regexp
func MatchesRegex(s string, pattern string) bool {
	obj, err := regexp.Match(pattern, []byte(s))
	return obj && err == nil
}

// GetImageVersion returns the version part of an image string
func GetImageVersion(imageVersionString string) string {
	// Find position of first "@", remove it and everything after it
	if strings.Contains(imageVersionString, "@") {
		imageVersionString = strings.Split(imageVersionString, "@")[0]
		imageVersionString = imageVersionString + ":unknown"
	}

	// If no version is given, assume "latest"
	if !strings.Contains(imageVersionString, ":") {
		return "latest"
	}

	// Split image string into parts
	parts := strings.Split(imageVersionString, ":")

	versionPart := parts[1]

	// Replace "+" with "-" in version text if version includes one
	versionPart = strings.ReplaceAll(versionPart, "+", "-")

	// Limit label-value to 63 characters
	if len(versionPart) > LabelValueMaxLength {
		versionPart = versionPart[:LabelValueMaxLength]
	}

	// While first character is not part of regex [a-z0-9A-Z] then remove it
	for len(versionPart) > 0 && !MatchesRegex(versionPart[:1], "[a-zA-Z0-9]") {
		versionPart = versionPart[1:]
	}

	return versionPart
}
