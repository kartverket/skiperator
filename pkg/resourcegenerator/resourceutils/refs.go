package resourceutils

import (
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func SetOwnerReference(skiperatorObject client.Object, obj client.Object, scheme *runtime.Scheme) error {
	if skiperatorObject.GetNamespace() != obj.GetNamespace() {
		return nil
	}
	return ctrlutil.SetControllerReference(skiperatorObject, obj, scheme)
}
