package common

import (
	"fmt"

	"github.com/kartverket/skiperator/api/common/istiotypes"
	"github.com/kartverket/skiperator/api/common/podtypes"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// +kubebuilder:object:generate=false
type SKIPObject interface {
	client.Object
	GetStatus() *SkiperatorStatus
	SetStatus(status SkiperatorStatus)
	GetDefaultLabels() map[string]string
	GetCommonSpec() *CommonSpec
}

var ErrNoGVK = fmt.Errorf("no GroupVersionKind found in the resources, cannot process resources")

// +kubebuilder:object:generate=true
// CommonSpec TODO: This needs some more thought. We should probably try to expand on it. v1Alpha2?
type CommonSpec struct {
	AccessPolicy  *podtypes.AccessPolicy
	GCP           *podtypes.GCP
	IstioSettings *istiotypes.IstioSettingsBase
	Image         string
}
