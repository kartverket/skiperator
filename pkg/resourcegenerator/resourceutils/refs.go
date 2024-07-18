package resourceutils

import (
	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func SetOwnerReference(app *skiperatorv1alpha1.Application, obj client.Object, scheme *runtime.Scheme) error {
	switch obj.(type) {
	//Certificates are created in istio-gateways namespace, so we cannot set ownerref
	case *certmanagerv1.Certificate:
		return nil
	default:
		if err := ctrlutil.SetControllerReference(app, obj, scheme); err != nil {
			return err
		}
	}
	return nil
}
