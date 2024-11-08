package resourceutils

import (
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"regexp"
	"strings"
)

func matchesRegex(s string, pattern string) bool {
	obj, err := regexp.Match(pattern, []byte(s))
	return obj && err == nil
}

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

// HumanReadableVersion returns the version part of an image string
func HumanReadableVersion(imageReference string) string {
	const LabelValueMaxLength = 63

	var allowedChars = regexp.MustCompile(`[A-Za-z0-9_.-]`)

	// Find position of first "@", remove it and everything after it
	if strings.Contains(imageReference, "@") {
		imageReference = strings.Split(imageReference, "@")[0]
	}

	lastColonPos := strings.LastIndex(imageReference, ":")
	if lastColonPos == -1 || lastColonPos == len(imageReference)-1 {
		return "latest"
	}
	versionPart := imageReference[lastColonPos+1:]
	imageReference = imageReference[:lastColonPos]

	// While first character is not part of regex [a-z0-9A-Z] then remove it
	for len(versionPart) > 0 && !matchesRegex(versionPart[:1], "[a-zA-Z0-9]") {
		versionPart = versionPart[1:]
	}

	// For each character in versionPart, replace characters that are not allowed in label-value
	var result strings.Builder
	for _, c := range versionPart {
		if allowedChars.MatchString(string(c)) {
			result.WriteRune(c)
		} else {
			result.WriteRune('-')
		}
	}
	versionPart = result.String()

	// Limit label-value to 63 characters
	if len(versionPart) > LabelValueMaxLength {
		versionPart = versionPart[:LabelValueMaxLength]
	}

	return versionPart
}
