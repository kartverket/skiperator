package sidecar

import (
	"github.com/kartverket/skiperator/pkg/reconciliation"
	networkingv1api "istio.io/api/networking/v1"
	networkingv1 "istio.io/client-go/pkg/apis/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func init() {
	multiGenerator.Register(reconciliation.NamespaceType, generateForNamespace)
}

func generateForNamespace(r reconciliation.Reconciliation) error {
	sidecar := networkingv1.Sidecar{ObjectMeta: metav1.ObjectMeta{Namespace: r.GetSKIPObject().GetName(), Name: "sidecar"}}

	sidecar.Spec = networkingv1api.Sidecar{
		OutboundTrafficPolicy: &networkingv1api.OutboundTrafficPolicy{
			Mode: networkingv1api.OutboundTrafficPolicy_REGISTRY_ONLY,
		},
	}

	r.AddResource(&sidecar)
	return nil
}
