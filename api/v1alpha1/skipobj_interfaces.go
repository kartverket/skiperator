package v1alpha1

import (
	"fmt"

	"github.com/kartverket/skiperator/api/v1alpha1/istiotypes"
	"github.com/kartverket/skiperator/api/v1alpha1/podtypes"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type SKIPObject interface {
	client.Object
	GetStatus() *SkiperatorStatus
	SetStatus(status SkiperatorStatus)
	GetDefaultLabels() map[string]string
	GetCommonSpec() *CommonSpec
	GetWorkloadName() string
}

var ErrNoGVK = fmt.Errorf("no GroupVersionKind found in the resources, cannot process resources")

// CommonSpec TODO: This needs some more thought. We should probably try to expand on it. v1Alpha2?
type CommonSpec struct {
	AccessPolicy  *podtypes.AccessPolicy
	GCP           *podtypes.GCP
	IstioSettings *istiotypes.IstioSettings
	Image         string
}
