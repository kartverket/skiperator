package v1alpha1

import (
	"fmt"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type SKIPObject interface {
	client.Object
	GetStatus() *SkiperatorStatus
	SetStatus(status SkiperatorStatus)
}

var ErrNoGVK error = fmt.Errorf("no GroupVersionKind found in the resources, cannot process resources")
