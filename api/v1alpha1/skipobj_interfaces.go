package v1alpha1

import (
	"fmt"
	"github.com/kartverket/skiperator/api/v1alpha1/podtypes"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

type SKIPObject interface {
	client.Object
	GetStatus() *SkiperatorStatus
	SetStatus(status SkiperatorStatus)
	GetDefaultLabels() map[string]string
	GetCommonSpec() *CommonSpec
}

var ErrNoGVK = fmt.Errorf("no GroupVersionKind found in the resources, cannot process resources")

// CommonSpec TODO: This needs some more thought. We should probably try to expand on it. v1Alpha2?
type CommonSpec struct {
	AccessPolicy *podtypes.AccessPolicy
	GCP          *podtypes.GCP
}

func getVersionLabel(imageVersionString string) string {
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
		return versionPart + "+" + parts[2]
	}
	return versionPart
}
