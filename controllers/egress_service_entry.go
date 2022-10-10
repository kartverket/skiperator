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
//+kubebuilder:rbac:groups=networking.istio.io,resources=serviceentries,verbs=get;list;watch;create;update;patch;delete

type EgressServiceEntryReconciler struct {
	client client.Client
	scheme *runtime.Scheme
}

func (r *EgressServiceEntryReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.client = mgr.GetClient()
	r.scheme = mgr.GetScheme()

	return ctrl.NewControllerManagedBy(mgr).
		For(&skiperatorv1alpha1.Application{}).
		Owns(&networkingv1beta1.ServiceEntry{}, builder.WithPredicates(
			matchesPredicate[*networkingv1beta1.ServiceEntry](isEgressServiceEntry),
		)).
		Complete(r)
}

func (r *EgressServiceEntryReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	application := skiperatorv1alpha1.Application{}
	err := r.client.Get(ctx, req.NamespacedName, &application)
	if err != nil {
		err = client.IgnoreNotFound(err)
		return reconcile.Result{}, err
	}
	application.FillDefaults()

	// Keep track of active service entries
	active := make(map[string]struct{}, len(application.Spec.AccessPolicy.Outbound.External))

	for _, rule := range application.Spec.AccessPolicy.Outbound.External {
		// Generate service entry name
		hash := fnv.New64()
		_, _ = hash.Write([]byte(rule.Host))
		name := fmt.Sprintf("%s-egress-%x", req.Name, hash.Sum64())
		active[name] = struct{}{}

		serviceEntry := networkingv1beta1.ServiceEntry{ObjectMeta: metav1.ObjectMeta{Namespace: req.Namespace, Name: name}}
		_, err = ctrlutil.CreateOrPatch(ctx, r.client, &serviceEntry, func() error {
			// Set application as owner of the service entry
			err = ctrlutil.SetControllerReference(&application, &serviceEntry, r.scheme)
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

			serviceEntry.Spec.Ports = make([]*networkingv1beta1api.Port, len(rule.Ports))
			for i, port := range rule.Ports {
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
	err = r.client.List(ctx, &serviceEntries, client.InNamespace(req.Namespace))
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
