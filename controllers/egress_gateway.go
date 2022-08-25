package controllers

import (
	"context"
	"fmt"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"hash/fnv"
	networkingv1beta1api "istio.io/api/networking/v1beta1"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"strings"
)

//+kubebuilder:rbac:groups=skiperator.kartverket.no,resources=applications,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=networking.istio.io,resources=gateways,verbs=get;list;watch;create;update;patch;delete

type EgressGatewayReconciler struct {
	client client.Client
	scheme *runtime.Scheme
}

func (r *EgressGatewayReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.client = mgr.GetClient()
	r.scheme = mgr.GetScheme()

	return ctrl.NewControllerManagedBy(mgr).
		For(&skiperatorv1alpha1.Application{}).
		Owns(&networkingv1beta1.Gateway{}, builder.WithPredicates(
			matchesPredicate[*networkingv1beta1.Gateway](isEgressGateway),
		)).
		Complete(r)
}

func (r *EgressGatewayReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	// Fetch application and fill defaults
	application := skiperatorv1alpha1.Application{}
	err := r.client.Get(ctx, req.NamespacedName, &application)
	if err != nil {
		err = client.IgnoreNotFound(err)
		return reconcile.Result{}, err
	}
	application.FillDefaults()

	// Keep track of active gateways
	active := make(map[string]struct{}, len(application.Spec.AccessPolicy.Outbound.External))

	// Generate separate gateway for each external rule
	for _, rule := range application.Spec.AccessPolicy.Outbound.External {
		// Generate gateway name
		hash := fnv.New64()
		_, _ = hash.Write([]byte(rule.Host))
		name := fmt.Sprintf("%s-egress-%x", req.Name, hash.Sum64())
		active[name] = struct{}{}

		gateway := networkingv1beta1.Gateway{ObjectMeta: metav1.ObjectMeta{Namespace: req.Namespace, Name: name}}
		_, err = ctrlutil.CreateOrPatch(ctx, r.client, &gateway, func() error {
			// Set application as owner of the gateway
			err = ctrlutil.SetControllerReference(&application, &gateway, r.scheme)
			if err != nil {
				return err
			}

			// Assign to external Envoy cluster
			gateway.Spec.Selector = map[string]string{"egress": "external"}

			// Generate separate destination per port
			gateway.Spec.Servers = make([]*networkingv1beta1api.Server, 0, len(rule.Ports))
			for _, port := range rule.Ports {
				server := &networkingv1beta1api.Server{}
				gateway.Spec.Servers = append(gateway.Spec.Servers, server)

				server.Hosts = []string{rule.Host}
				server.Port = &networkingv1beta1api.Port{}
				server.Port.Name = port.Name
				server.Port.Number = uint32(port.Port)
				server.Port.Protocol = port.Protocol

				if port.Protocol == "HTTPS" {
					server.Tls = &networkingv1beta1api.ServerTLSSettings{}
					server.Tls.Mode = networkingv1beta1api.ServerTLSSettings_PASSTHROUGH
				}
			}

			return nil
		})
		if err != nil {
			return reconcile.Result{}, err
		}
	}

	// Clear out unused gateways
	gateways := networkingv1beta1.GatewayList{}
	err = r.client.List(ctx, &gateways, client.InNamespace(req.Namespace))
	if err != nil {
		return reconcile.Result{}, err
	}

	for i := range gateways.Items {
		gateway := &gateways.Items[i]

		// Skip unrelated gateways
		if !isEgressGateway(gateway) {
			continue
		}

		// Skip active gateways
		_, ok := active[gateway.Name]
		if ok {
			continue
		}

		// Delete the rest
		err = r.client.Delete(ctx, gateway)
		err = client.IgnoreNotFound(err)
		if err != nil {
			return reconcile.Result{}, err
		}
	}

	return reconcile.Result{}, nil
}

// Filter for gateways named like *-egress-*
func isEgressGateway(gateway *networkingv1beta1.Gateway) bool {
	segments := strings.Split(gateway.Name, "-")

	if len(segments) != 3 {
		return false
	}

	return segments[1] == "egress"
}
