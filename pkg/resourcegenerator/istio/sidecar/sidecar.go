package sidecar

import (
	"fmt"
	"github.com/kartverket/skiperator/pkg/reconciliation"
	networkingv1beta1api "istio.io/api/networking/v1beta1"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TODO investigate: this doesn't seem to be doing anything on the cluster today?
func Generate(r reconciliation.Reconciliation) error {
	ctxLog := r.GetLogger()
	ctxLog.Debug("Attempting to generate istio sidecar resource for namespace", "namespace", r.GetSKIPObject().GetName())

	if r.GetType() != reconciliation.NamespaceType {
		return fmt.Errorf(" istio sidecar resource supports namespace type")
	}

	sidecar := networkingv1beta1.Sidecar{ObjectMeta: metav1.ObjectMeta{Namespace: r.GetSKIPObject().GetName(), Name: "sidecar"}}

	sidecar.Spec = networkingv1beta1api.Sidecar{
		OutboundTrafficPolicy: &networkingv1beta1api.OutboundTrafficPolicy{
			Mode: networkingv1beta1api.OutboundTrafficPolicy_REGISTRY_ONLY,
		},
	}

	r.AddResource(&sidecar)

	ctxLog.Debug("Finished generating default deny network policy for namespace", "namespace", r.GetSKIPObject().GetName())
	return nil
}
