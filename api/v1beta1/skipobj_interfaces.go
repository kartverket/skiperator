package v1beta1

import (
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"
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
	AccessPolicy  *AccessPolicy
	GCP           *GCP
	IstioSettings *IstioSettingsBase
	Image         string
}
