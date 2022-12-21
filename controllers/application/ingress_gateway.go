package applicationcontroller

import (
	"context"
	"fmt"
	"hash/fnv"
	"strings"

	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	util "github.com/kartverket/skiperator/pkg/util"
	"golang.org/x/exp/slices"
	networkingv1beta1api "istio.io/api/networking/v1beta1"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *ApplicationReconciler) reconcileIngressGateway(ctx context.Context, application *skiperatorv1alpha1.Application) (reconcile.Result, error) {
	controllerName := "IngressGateway"
	r.SetControllerProgressing(ctx, application, controllerName)

	// Generate separate gateway for each ingress
	for _, hostname := range application.Spec.Ingresses {

		name := fmt.Sprintf("%s-ingress-%x", application.Name, generateHostnameHash(hostname))

		gateway := networkingv1beta1.Gateway{ObjectMeta: metav1.ObjectMeta{Namespace: application.Namespace, Name: name}}
		_, err := ctrlutil.CreateOrPatch(ctx, r.GetClient(), &gateway, func() error {
			// Set application as owner of the gateway
			err := ctrlutil.SetControllerReference(application, &gateway, r.GetScheme())
			if err != nil {
				r.SetControllerError(ctx, application, controllerName, err)
				return err
			}

			if util.IsInternal(hostname) {
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
			gateway.Spec.Servers[1].Tls.CredentialName = application.Namespace + "-" + name

			return nil
		})
		if err != nil {
			r.SetControllerError(ctx, application, controllerName, err)
			return reconcile.Result{}, err
		}
	}

	// Clear out unused gateways
	gateways := networkingv1beta1.GatewayList{}
	err := r.GetClient().List(ctx, &gateways, client.InNamespace(application.Namespace))

	if err != nil {
		r.SetControllerError(ctx, application, controllerName, err)
		return reconcile.Result{}, err
	}

	for _, gateway := range gateways.Items {
		// Skip unrelated gateways
		if !isIngressGateway(gateway) {
			continue
		}

		// Skip gateways in which the reconciled application is not the owner
		applicationOwnerIndex := slices.IndexFunc(gateway.GetOwnerReferences(), func(ownerReference metav1.OwnerReference) bool {
			return ownerReference.Name == application.Name
		})
		gatewayOwnedByThisApplication := applicationOwnerIndex != -1
		if !gatewayOwnedByThisApplication {
			continue
		}

		ingressGatewayInApplicationSpecIndex := slices.IndexFunc(application.Spec.Ingresses, func(hostname string) bool {
			ingressName := fmt.Sprintf("%s-ingress-%x", application.Name, generateHostnameHash(hostname))
			return gateway.Name == ingressName
		})
		ingressGatewayInApplicationSpec := ingressGatewayInApplicationSpecIndex != -1
		if ingressGatewayInApplicationSpec {
			continue
		}

		// Delete the rest
		err = r.GetClient().Delete(ctx, gateway)
		err = client.IgnoreNotFound(err)
		if err != nil {
			r.SetControllerError(ctx, application, controllerName, err)
			return reconcile.Result{}, err
		}
	}

	r.SetControllerFinishedOutcome(ctx, application, controllerName, err)

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

func generateHostnameHash(hostname string) uint64 {
	hash := fnv.New64()
	_, _ = hash.Write([]byte(hostname))
	return hash.Sum64()
}
