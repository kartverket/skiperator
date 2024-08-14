package resourceutils

import (
	"fmt"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

func AddGVK(scheme *runtime.Scheme, obj client.Object) error {
	gvk, err := apiutil.GVKForObject(obj, scheme)
	if err != nil {
		return fmt.Errorf("failed to get GVK for object, need gvk to proceed. type may not be added to schema: %w", err)
	}
	obj.GetObjectKind().SetGroupVersionKind(gvk)
	return nil
}
