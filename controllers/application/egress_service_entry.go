package applicationcontroller

import (
	"context"
	"fmt"
	"github.com/kartverket/skiperator/api/v1alpha1/podtypes"
	"regexp"

	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/util"
	"golang.org/x/exp/slices"
	networkingv1beta1api "istio.io/api/networking/v1beta1"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *ApplicationReconciler) reconcileEgressServiceEntry(ctx context.Context, application *skiperatorv1alpha1.Application) (reconcile.Result, error) {
	controllerName := "EgressServiceEntry"
	r.SetControllerProgressing(ctx, application, controllerName)

	if application.Spec.AccessPolicy != nil {
		for _, rule := range (*application.Spec.AccessPolicy).Outbound.External {
			name := fmt.Sprintf("%s-egress-%x", application.Name, util.GenerateHashFromName(rule.Host))

			serviceEntry := networkingv1beta1.ServiceEntry{ObjectMeta: metav1.ObjectMeta{Namespace: application.Namespace, Name: name}}
			_, err := ctrlutil.CreateOrPatch(ctx, r.GetClient(), &serviceEntry, func() error {
				// Set application as owner of the service entry
				err := ctrlutil.SetControllerReference(application, &serviceEntry, r.GetScheme())
				if err != nil {
					r.SetControllerError(ctx, application, controllerName, err)
					return err
				}
				r.SetLabelsFromApplication(ctx, &serviceEntry, *application)
				util.SetCommonAnnotations(&serviceEntry)

				resolution, adresses, endpoints := getIpData(rule.Ip)

				serviceEntry.Spec = networkingv1beta1api.ServiceEntry{
					// Avoid leaking service entry to other namespaces
					ExportTo:   []string{".", "istio-system", "istio-gateways"},
					Hosts:      []string{rule.Host},
					Resolution: resolution,
					Addresses:  adresses,
					Endpoints:  endpoints,
					Ports:      r.getPorts(rule.Ports, rule.Ip, *application),
				}

				return nil
			})
			if err != nil {
				r.SetControllerError(ctx, application, controllerName, err)
				return reconcile.Result{}, err
			}
		}
	}

	// Clear out unused service entries
	serviceEntries := networkingv1beta1.ServiceEntryList{}
	err := r.GetClient().List(ctx, &serviceEntries, client.InNamespace(application.Namespace))
	if err != nil {
		r.SetControllerError(ctx, application, controllerName, err)
		return reconcile.Result{}, err
	}

	for _, serviceEntry := range serviceEntries.Items {
		// Skip unrelated service entries
		if !isEgressServiceEntry(serviceEntry) {
			continue
		}

		applicationOwnerIndex := slices.IndexFunc(serviceEntry.GetOwnerReferences(), func(ownerReference metav1.OwnerReference) bool {
			return ownerReference.Name == application.Name
		})
		serviceEntryOwnedByThisApplication := applicationOwnerIndex != -1
		if !serviceEntryOwnedByThisApplication {
			continue
		}

		serviceEntryInApplicationSpecIndex := slices.IndexFunc(application.Spec.AccessPolicy.Outbound.External, func(rule podtypes.ExternalRule) bool {
			egressName := fmt.Sprintf("%s-egress-%x", application.Name, util.GenerateHashFromName(rule.Host))
			return serviceEntry.Name == egressName
		})
		ingressGatewayInApplicationSpec := serviceEntryInApplicationSpecIndex != -1
		if ingressGatewayInApplicationSpec {
			continue
		}

		// Delete the rest
		err = r.GetClient().Delete(ctx, serviceEntry)
		err = client.IgnoreNotFound(err)
		if err != nil {
			r.SetControllerError(ctx, application, controllerName, err)
			return reconcile.Result{}, err
		}
	}

	r.SetControllerFinishedOutcome(ctx, application, controllerName, err)

	return reconcile.Result{}, err
}

func (r ApplicationReconciler) getPorts(externalPorts []podtypes.ExternalPort, ruleIP string, application skiperatorv1alpha1.Application) []*networkingv1beta1api.Port {
	ports := []*networkingv1beta1api.Port{}

	if len(externalPorts) == 0 {
		ports = append(ports, &networkingv1beta1api.Port{
			Name:     "https",
			Number:   uint32(443),
			Protocol: "HTTPS",
		})

		return ports
	}

	for _, port := range externalPorts {
		if ruleIP == "" && port.Protocol == "TCP" {
			r.GetRecorder().Eventf(
				&application,
				corev1.EventTypeWarning, "Invalid",
				"A static IP must be set for TCP port %d",
				port.Port,
			)
			continue
		}

		ports = append(ports, &networkingv1beta1api.Port{
			Name:     port.Name,
			Number:   uint32(port.Port),
			Protocol: port.Protocol,
		})

	}

	return ports
}

func getIpData(ip string) (networkingv1beta1api.ServiceEntry_Resolution, []string, []*networkingv1beta1api.WorkloadEntry) {
	if ip == "" {
		return networkingv1beta1api.ServiceEntry_DNS, nil, nil
	}

	return networkingv1beta1api.ServiceEntry_STATIC, []string{ip}, []*networkingv1beta1api.WorkloadEntry{{Address: ip}}
}

// Filter for service entries named like *-egress-*
func isEgressServiceEntry(serviceEntry *networkingv1beta1.ServiceEntry) bool {
	match, _ := regexp.MatchString("^.*-egress-.*$", serviceEntry.Name)

	return match
}
