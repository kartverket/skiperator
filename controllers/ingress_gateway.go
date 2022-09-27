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

type IngressGatewayReconciler struct {
	client client.Client
	scheme *runtime.Scheme
}

func (r *IngressGatewayReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.client = mgr.GetClient()
	r.scheme = mgr.GetScheme()

	return ctrl.NewControllerManagedBy(mgr).
		For(&skiperatorv1alpha1.Application{}).
		Owns(&networkingv1beta1.Gateway{}, builder.WithPredicates(
			matchesPredicate[*networkingv1beta1.Gateway](isIngressGateway),
		)).
		Complete(r)
}

func (r *IngressGatewayReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	// Fetch application and fill defaults
	application := skiperatorv1alpha1.Application{}
	err := r.client.Get(ctx, req.NamespacedName, &application)
	if err != nil {
		err = client.IgnoreNotFound(err)
		return reconcile.Result{}, err
	}
	application.FillDefaults()

	// Keep track of active gateways
	active := make(map[string]struct{}, len(application.Spec.Ingresses))

	// Generate separate gateway for each ingress
	for _, hostname := range application.Spec.Ingresses {
		// Generate gateway name
		hash := fnv.New64()
		_, _ = hash.Write([]byte(hostname))
		name := fmt.Sprintf("%s-ingress-%x", req.Name, hash.Sum64())
		active[name] = struct{}{}

		gateway := networkingv1beta1.Gateway{ObjectMeta: metav1.ObjectMeta{Namespace: req.Namespace, Name: name}}
		_, err = ctrlutil.CreateOrPatch(ctx, r.client, &gateway, func() error {
			// Set application as owner of the gateway
			err = ctrlutil.SetControllerReference(&application, &gateway, r.scheme)
			if err != nil {
				return err
			}

			if isInternal(hostname) {
				gateway.Spec.Selector = map[string]string{"ingress": "internal"}
			} else {
				gateway.Spec.Selector = map[string]string{"ingress": "external"}
			}

			gateway.Spec.Servers = make([]*networkingv1beta1api.Server, 2)

			gateway.Spec.Servers[0] = &networkingv1beta1api.Server{}
			gateway.Spec.Servers[0].Hosts = []string{hostname}
			gateway.Spec.Servers[0].Port = &networkingv1beta1api.Port{}
			gateway.Spec.Servers[0].Port.Number = 80
			gateway.Spec.Servers[0].Port.Name = "http"
			gateway.Spec.Servers[0].Port.Protocol = "HTTP"

			gateway.Spec.Servers[1] = &networkingv1beta1api.Server{}
			gateway.Spec.Servers[1].Hosts = []string{hostname}
			gateway.Spec.Servers[1].Port = &networkingv1beta1api.Port{}
			gateway.Spec.Servers[1].Port.Number = 443
			gateway.Spec.Servers[1].Port.Name = "https"
			gateway.Spec.Servers[1].Port.Protocol = "HTTPS"
			gateway.Spec.Servers[1].Tls = &networkingv1beta1api.ServerTLSSettings{}
			gateway.Spec.Servers[1].Tls.Mode = networkingv1beta1api.ServerTLSSettings_SIMPLE
			gateway.Spec.Servers[1].Tls.CredentialName = req.Namespace + "-" + name

			return nil
		})
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
		if !isIngressGateway(gateway) {
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

	return reconcile.Result{}, err
}

// Filter for gateways named like *-ingress
func isIngressGateway(gateway *networkingv1beta1.Gateway) bool {
	segments := strings.Split(gateway.Name, "-")

	if len(segments) != 3 {
		return false
	}

	return segments[1] == "ingress"
}
