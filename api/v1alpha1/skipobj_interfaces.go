package v1alpha1

import (
	"fmt"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type SKIPObject interface {
	client.Object
	GetStatus() *SkiperatorStatus
	SetStatus(status SkiperatorStatus)
	GetDefaultLabels() map[string]string
}

var ErrNoGVK = fmt.Errorf("no GroupVersionKind found in the resources, cannot process resources")
