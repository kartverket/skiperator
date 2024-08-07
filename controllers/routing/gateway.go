package routingcontroller

import (
	"context"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/util"
	networkingv1beta1api "istio.io/api/networking/v1beta1"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *RoutingReconciler) reconcileGateway(ctx context.Context, routing *skiperatorv1alpha1.Routing) (reconcile.Result, error) {
	h, err := routing.Spec.GetHost()
	if err != nil {
		err = r.setConditionGatewaySynced(ctx, routing, ConditionStatusFalse, err.Error())
		return util.DoNotRequeue()
	}

	gateway := networkingv1beta1.Gateway{ObjectMeta: metav1.ObjectMeta{Namespace: routing.Namespace, Name: routing.GetGatewayName()}}

	var determinedCredentialName string
	if h.UsesCustomCert() {
		determinedCredentialName = *h.CustomCertificateSecret
	} else {
		determinedCredentialName, err = routing.GetCertificateName()
		if err != nil {
			err = r.setConditionGatewaySynced(ctx, routing, ConditionStatusFalse, err.Error())
			return util.RequeueWithError(err)
		}
	}

	_, err = ctrlutil.CreateOrPatch(ctx, r.GetClient(), &gateway, func() error {
		err := ctrlutil.SetControllerReference(routing, &gateway, r.GetScheme())
		if err != nil {
			return err
		}

		util.SetCommonAnnotations(&gateway)

		gateway.Spec.Selector = util.GetIstioGatewayLabelSelector(h.Hostname)
		gateway.Spec.Servers = []*networkingv1beta1api.Server{
			{
				Hosts: []string{h.Hostname},
				Port: &networkingv1beta1api.Port{
					Number:   80,
					Name:     "http",
					Protocol: "HTTP",
				},
			},
			{
				Hosts: []string{h.Hostname},
				Port: &networkingv1beta1api.Port{
					Number:   443,
					Name:     "https",
					Protocol: "HTTPS",
				},
				Tls: &networkingv1beta1api.ServerTLSSettings{
					Mode:           networkingv1beta1api.ServerTLSSettings_SIMPLE,
					CredentialName: determinedCredentialName,
				},
			},
		}

		return nil
	})
	if err != nil {
		err = r.setConditionGatewaySynced(ctx, routing, ConditionStatusFalse, err.Error())
		return util.RequeueWithError(err)
	}

	m := ConditionMessageGatewaySynced
	if h.UsesCustomCert() {
		m = ConditionMessageGatewaySyncedCustomCertificate
	}
	err = r.setConditionGatewaySynced(ctx, routing, ConditionStatusTrue, m)
	return util.RequeueWithError(err)

}
