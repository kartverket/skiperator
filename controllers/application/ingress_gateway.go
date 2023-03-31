package applicationcontroller

import (
	"context"
	"fmt"
	"regexp"

	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/util"
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

		name := fmt.Sprintf("%s-ingress-%x", application.Name, util.GenerateHashFromName(hostname))

		gateway := networkingv1beta1.Gateway{ObjectMeta: metav1.ObjectMeta{Namespace: application.Namespace, Name: name}}
		_, err := ctrlutil.CreateOrPatch(ctx, r.GetClient(), &gateway, func() error {
			// Set application as owner of the gateway
			err := ctrlutil.SetControllerReference(application, &gateway, r.GetScheme())
			if err != nil {
				r.SetControllerError(ctx, application, controllerName, err)
				return err
			}

			r.SetLabelsFromApplication(ctx, &gateway, *application)
			util.SetCommonAnnotations(&gateway)

			internalSelector := "istio-ingress-internal"
			externalSelector := "istio-ingress-external"

			if util.IsInternal(hostname) {
				gateway.Spec.Selector = map[string]string{"app": internalSelector}
			} else {
				gateway.Spec.Selector = map[string]string{"app": externalSelector}
			}

			gatewayServersToAdd := []*networkingv1beta1api.Server{}

			baseHttpGatewayServer := &networkingv1beta1api.Server{
				Hosts: []string{hostname},
				Port: &networkingv1beta1api.Port{
					Number:   80,
					Name:     "http",
					Protocol: "HTTP",
				},
			}

			if application.Spec.RedirectIngresses {
				wellKnownGatewayServer := &networkingv1beta1api.Server{
					Hosts: []string{hostname + "/.well-known/acme-challenge"},
					Port: &networkingv1beta1api.Port{
						Number:   80,
						Name:     "http",
						Protocol: "HTTP",
					},
				}

				baseHttpGatewayServer.Tls = &networkingv1beta1api.ServerTLSSettings{
					HttpsRedirect: true,
				}

				gatewayServersToAdd = append(gatewayServersToAdd, wellKnownGatewayServer)
			}

			httpsGatewayServer := &networkingv1beta1api.Server{
				Hosts: []string{hostname},
				Port: &networkingv1beta1api.Port{
					Number:   443,
					Name:     "https",
					Protocol: "HTTPS",
				},
				Tls: &networkingv1beta1api.ServerTLSSettings{
					Mode:           networkingv1beta1api.ServerTLSSettings_SIMPLE,
					CredentialName: application.Namespace + "-" + name,
				},
			}

			gatewayServersToAdd = append(gatewayServersToAdd, baseHttpGatewayServer, httpsGatewayServer)

			gateway.Spec.Servers = gatewayServersToAdd

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

		applicationOwnerIndex := slices.IndexFunc(gateway.GetOwnerReferences(), func(ownerReference metav1.OwnerReference) bool {
			return ownerReference.Name == application.Name
		})
		gatewayOwnedByThisApplication := applicationOwnerIndex != -1
		if !gatewayOwnedByThisApplication {
			continue
		}

		ingressGatewayInApplicationSpecIndex := slices.IndexFunc(application.Spec.Ingresses, func(hostname string) bool {
			ingressName := fmt.Sprintf("%s-ingress-%x", application.Name, util.GenerateHashFromName(hostname))
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

// Filter for gateways named like *-ingress-*
func isIngressGateway(gateway *networkingv1beta1.Gateway) bool {
	match, _ := regexp.MatchString("^.*-ingress-.*$", gateway.Name)

	return match
}
