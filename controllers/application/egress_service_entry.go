package applicationcontroller

import (
	"context"
	"fmt"
	"regexp"

	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	util "github.com/kartverket/skiperator/pkg/util"
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

	for _, rule := range application.Spec.AccessPolicy.Outbound.External {
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
			util.SetCommonAnnotations(ctx, &serviceEntry)

			// Avoid leaking service entry to other namespaces
			serviceEntry.Spec.ExportTo = []string{".", "istio-system"}

			serviceEntry.Spec.Hosts = []string{rule.Host}
			if rule.Ip == "" {
				serviceEntry.Spec.Resolution = networkingv1beta1api.ServiceEntry_DNS
			} else {
				serviceEntry.Spec.Resolution = networkingv1beta1api.ServiceEntry_STATIC
				serviceEntry.Spec.Addresses = []string{rule.Ip}
				serviceEntry.Spec.Endpoints = []*networkingv1beta1api.WorkloadEntry{{Address: rule.Ip}}
			}

			ports := rule.Ports

			// When not specified default to opening HTTPS
			if len(ports) == 0 {
				ports = make([]skiperatorv1alpha1.ExternalPort, 1)

				ports[0].Name = "https"
				ports[0].Port = 443
				ports[0].Protocol = "HTTPS"
			}

			serviceEntry.Spec.Ports = make([]*networkingv1beta1api.Port, len(ports))
			for i, port := range ports {
				if rule.Ip == "" && port.Protocol == "TCP" {
					r.GetRecorder().Eventf(
						application,
						corev1.EventTypeWarning, "Invalid",
						"A static IP must be set for TCP port %d",
						port.Port,
					)
					continue
				}

				serviceEntry.Spec.Ports[i] = &networkingv1beta1api.Port{}
				serviceEntry.Spec.Ports[i].Name = port.Name
				serviceEntry.Spec.Ports[i].Number = uint32(port.Port)
				serviceEntry.Spec.Ports[i].Protocol = port.Protocol
			}

			return nil
		})
		if err != nil {
			r.SetControllerError(ctx, application, controllerName, err)
			return reconcile.Result{}, err
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

		serviceEntryInApplicationSpecIndex := slices.IndexFunc(application.Spec.AccessPolicy.Outbound.External, func(rule skiperatorv1alpha1.ExternalRule) bool {
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

// Filter for service entries named like *-egress-*
func isEgressServiceEntry(serviceEntry *networkingv1beta1.ServiceEntry) bool {
	match, _ := regexp.MatchString("^.*-egress-.*$", serviceEntry.Name)

	return match
}
