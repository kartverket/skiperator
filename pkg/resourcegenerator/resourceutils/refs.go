package resourceutils

import (
	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func SetOwnerReference(skiperatorObject client.Object, obj client.Object, scheme *runtime.Scheme) error {
	switch obj.(type) {
	//Certificates are created in istio-gateways namespace, so we cannot set ownerref
	case *certmanagerv1.Certificate:
		return nil
	default:
		if err := ctrlutil.SetControllerReference(skiperatorObject, obj, scheme); err != nil {
			return err
		}
	}
	return nil
}
