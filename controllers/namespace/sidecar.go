package namespacecontroller

import (
	"context"

	networkingv1beta1api "istio.io/api/networking/v1beta1"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *NamespaceReconciler) reconcileSidecar(ctx context.Context, namespace *corev1.Namespace) (reconcile.Result, error) {
	sidecar := networkingv1beta1.Sidecar{ObjectMeta: metav1.ObjectMeta{Namespace: namespace.Name, Name: "sidecar"}}
	_, err := ctrlutil.CreateOrPatch(ctx, r.GetClient(), &sidecar, func() error {
		// Set namespace as owner of the sidecar
		err := ctrlutil.SetControllerReference(namespace, &sidecar, r.GetScheme())
		if err != nil {
			return err
		}

		sidecar.Spec = networkingv1beta1api.Sidecar{
			OutboundTrafficPolicy: &networkingv1beta1api.OutboundTrafficPolicy{
				Mode: networkingv1beta1api.OutboundTrafficPolicy_REGISTRY_ONLY,
			},
		}

		return nil
	})
	return reconcile.Result{}, err
}
