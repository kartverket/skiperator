package resourceutils

import (
	"regexp"
	"strings"

	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/log"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

const (
	LabelValueMaxLength = 63
	defaultImageVersion = "latest"
	unknownImageVersion = "err-unknown"
)

var (
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

// HumanReadableVersion parses an image reference (e.g. "ghcr.io/org/some-team/some-app:v1.2.3")
// and returns a human-readable version string. If the image reference is empty, it returns a default
// unknown version string. The function removes any digest part (everything after and including "@")
// from the image reference, extracts the version part (everything after the last ":"), trims non-alphanumeric
// prefix characters from the version part, replaces any characters not allowed in Kubernetes label values
// with a hyphen ("-"), and limits the version string to a maximum length of 63 characters. If the version
// part is not found, it returns a default version string.
func HumanReadableVersion(logger *log.Logger, imageReference string) string {
	logz := (*logger).GetLogger().WithValues("helper", "HumanReadableVersion")

	if len(imageReference) == 0 {
		logz.Info("imageReference had length 0")
		return unknownImageVersion
	}

	processedImageRef := strings.Clone(imageReference)

	// Find position of first "@", remove it and everything after it
	if strings.Contains(processedImageRef, "@") {
		processedImageRef = strings.Split(processedImageRef, "@")[0]
	}

	lastColonPos := strings.LastIndex(processedImageRef, ":")
	if lastColonPos == -1 || lastColonPos == len(processedImageRef)-1 {
		return defaultImageVersion
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
		logz.Info("Trimming version length because it too long", "ref", imageReference, "trimmedVersion", versionPart)
	}

	return versionPart
}
