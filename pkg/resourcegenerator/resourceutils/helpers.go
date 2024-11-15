package resourceutils

import (
	"regexp"
	"strings"

	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/log"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

const LabelValueMaxLength = 63

var (
	logger                  = log.NewLogger().WithName("resourceutils")
	allowedChars            = regexp.MustCompile(`[A-Za-z0-9_.-]`)
	allowedPrefixCharacters = regexp.MustCompile(`[a-zA-Z0-9]`)
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

// HumanReadableVersion returns the version part of an image string
func HumanReadableVersion(imageReference string) string {
	processedImageRef := strings.Clone(imageReference)

	// Find position of first "@", remove it and everything after it
	if strings.Contains(processedImageRef, "@") {
		processedImageRef = strings.Split(processedImageRef, "@")[0]
	}

	lastColonPos := strings.LastIndex(processedImageRef, ":")
	if lastColonPos == -1 || lastColonPos == len(processedImageRef)-1 {
		return "latest"
	}

	versionPart := processedImageRef[lastColonPos+1:]
	processedImageRef = processedImageRef[:lastColonPos]

	// Trim non-alphanumeric prefix
	versionPart = strings.TrimLeftFunc(versionPart, func(r rune) bool {
		return !allowedPrefixCharacters.MatchString(string(r))
	})

	// For each character in versionPart, replace characters that are not allowed in label-value
	versionPart = strings.Map(func(r rune) rune {
		if allowedChars.MatchString(string(r)) {
			return r
		}
		return '-'
	}, versionPart)

	// Limit label-value to 63 characters
	if len(versionPart) > LabelValueMaxLength {
		versionPart = versionPart[:LabelValueMaxLength]
		logger.Info("Trimming version length because it too long", "ref", imageReference, "trimmedVersion", versionPart)
	}

	return versionPart
}
