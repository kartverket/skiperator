package routingcontroller

import (
	"context"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/util"
	networkingv1beta1api "istio.io/api/networking/v1beta1"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *RoutingReconciler) reconcileGateway(ctx context.Context, routing *skiperatorv1alpha1.Routing) (reconcile.Result, error) {
	var err error

	name := routing.Name + "-ingress"
	gateway := networkingv1beta1.Gateway{ObjectMeta: metav1.ObjectMeta{Namespace: routing.Namespace, Name: name}}

	_, err = ctrlutil.CreateOrPatch(ctx, r.GetClient(), &gateway, func() error {
		// Set application as owner of the gateway
		err := ctrlutil.SetControllerReference(routing, &gateway, r.GetScheme())
		if err != nil {
			//r.SetControllerError(ctx, application, controllerName, err)
			return err
		}

		//r.SetLabelsFromApplication(&gateway, *application)
		util.SetCommonAnnotations(&gateway)

		gateway.Spec.Selector = util.GetIstioGatewayLabelSelector(routing.Spec.Hostname)
		gateway.Spec.Servers = []*networkingv1beta1api.Server{
			{
				Hosts: []string{routing.Spec.Hostname},
				Port: &networkingv1beta1api.Port{
					Number:   80,
					Name:     "http",
					Protocol: "HTTP",
				},
			},
			{
				Hosts: []string{routing.Spec.Hostname},
				Port: &networkingv1beta1api.Port{
					Number:   443,
					Name:     "https",
					Protocol: "HTTPS",
				},
				Tls: &networkingv1beta1api.ServerTLSSettings{
					Mode:           networkingv1beta1api.ServerTLSSettings_SIMPLE,
					CredentialName: util.GetGatewaySecretName(routing.Namespace, routing.Name),
				},
			},
		}

		return nil
	})
	if err != nil {
		//r.SetControllerError(ctx, application, controllerName, err)
		return util.RequeueWithError(err)
	}

	return ctrl.Result{}, nil

}
