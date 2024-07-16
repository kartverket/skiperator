package resourceutils

import (
	"fmt"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

func SetGVK(obj client.Object, scheme *runtime.Scheme) error {
	gvk, err := apiutil.GVKForObject(obj, scheme)
	if err != nil {
		return fmt.Errorf("error getting GVK for object: %w", err)
	}
	obj.GetObjectKind().SetGroupVersionKind(gvk)
	return nil
}
