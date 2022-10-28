package controllers

import (
	"context"
	"fmt"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"hash/fnv"
	networkingv1beta1api "istio.io/api/networking/v1beta1"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"strings"
)

//+kubebuilder:rbac:groups=skiperator.kartverket.no,resources=applications,verbs=get;list;watch
//+kubebuilder:rbac:groups=networking.istio.io,resources=serviceentries,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=events,verbs=create;patch

type EgressServiceEntryReconciler struct {
	client   client.Client
	scheme   *runtime.Scheme
	recorder record.EventRecorder
}

func (r *EgressServiceEntryReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.client = mgr.GetClient()
	r.scheme = mgr.GetScheme()
	r.recorder = mgr.GetEventRecorderFor("egress-serviceentry-controller")

	return newControllerManagedBy[*skiperatorv1alpha1.Application](mgr).
		For(&skiperatorv1alpha1.Application{}).
		Owns(&networkingv1beta1.ServiceEntry{}, builder.WithPredicates(
			matchesPredicate[*networkingv1beta1.ServiceEntry](isEgressServiceEntry),
		)).
		Complete(r)
}

func (r *EgressServiceEntryReconciler) Reconcile(ctx context.Context, application *skiperatorv1alpha1.Application) (reconcile.Result, error) {
	application.FillDefaults()

	// Keep track of active service entries
	active := make(map[string]struct{}, len(application.Spec.AccessPolicy.Outbound.External))

	for _, rule := range application.Spec.AccessPolicy.Outbound.External {
		// Generate service entry name
		hash := fnv.New64()
		_, _ = hash.Write([]byte(rule.Host))
		name := fmt.Sprintf("%s-egress-%x", application.Name, hash.Sum64())
		active[name] = struct{}{}

		serviceEntry := networkingv1beta1.ServiceEntry{ObjectMeta: metav1.ObjectMeta{Namespace: application.Namespace, Name: name}}
		_, err := ctrlutil.CreateOrPatch(ctx, r.client, &serviceEntry, func() error {
			// Set application as owner of the service entry
			err := ctrlutil.SetControllerReference(application, &serviceEntry, r.scheme)
			if err != nil {
				return err
			}

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

			// When not specified default to opening HTTP and HTTPS
			if len(ports) == 0 {
				ports = make([]*networkingv1beta1api.Port, 2)

				ports[0].Name = "http"
				ports[0].Number = 80
				ports[0].Protocol = "HTTPS"

				ports[1].Name = "https"
				ports[1].Number = 443
				ports[1].Protocol = "HTTPS"
			}

			serviceEntry.Spec.Ports = make([]*networkingv1beta1api.Port, len(ports))
			for i, port := range ports {
				if rule.Ip == "" && port.Protocol == "TCP" {
					r.recorder.Eventf(
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
			return reconcile.Result{}, err
		}
	}

	// Clear out unused service entries
	serviceEntries := networkingv1beta1.ServiceEntryList{}
	err := r.client.List(ctx, &serviceEntries, client.InNamespace(application.Namespace))
	if err != nil {
		return reconcile.Result{}, err
	}

	for i := range serviceEntries.Items {
		serviceEntry := serviceEntries.Items[i]

		// Skip unrelated service entries
		if !isEgressServiceEntry(serviceEntry) {
			continue
		}

		// Skip active service entries
		_, ok := active[serviceEntry.Name]
		if ok {
			continue
		}

		// Delete the rest
		err = r.client.Delete(ctx, serviceEntry)
		err = client.IgnoreNotFound(err)
		if err != nil {
			return reconcile.Result{}, err
		}
	}

	return reconcile.Result{}, nil
}

// Filter for service entries named like *-egress-*
func isEgressServiceEntry(serviceEntry *networkingv1beta1.ServiceEntry) bool {
	segments := strings.Split(serviceEntry.Name, "-")

	if len(segments) != 3 {
		return false
	}

	return segments[1] == "egress"
}
