// +groupName=skiperator.kartverket.no
// +versionName=v1alpha1
package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

//go:generate controller-gen object

var (
	GroupVersion  = schema.GroupVersion{Group: "skiperator.kartverket.no", Version: "v1alpha1"}
	schemeBuilder = &scheme.Builder{GroupVersion: GroupVersion}
	AddToScheme   = schemeBuilder.AddToScheme
)

func init() {
	schemeBuilder.Register(&ApplicationList{}, &Application{})
	schemeBuilder.Register(&SKIPJobList{}, &SKIPJob{})
	schemeBuilder.Register(&RoutingList{}, &Routing{})
}
