package routingcontroller

import (
	"context"
	"fmt"
	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/util"
	istionetworkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// +kubebuilder:rbac:groups=skiperator.kartverket.no,resources=routings;routings/status,verbs=get;list;watch;update
// +kubebuilder:rbac:groups=skiperator.kartverket.no,resources=applications;applications/status,verbs=get;list;watch
// +kubebuilder:rbac:groups=networking.k8s.io,resources=networkpolicies,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=networking.istio.io,resources=gateways;virtualservices,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cert-manager.io,resources=certificates,verbs=get;list;watch;create;update;patch;delete

type RoutingReconciler struct {
	util.ReconcilerBase
}

func (r *RoutingReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		// GenerationChangedPredicate is now only applied to the SkipJob itself to allow status changes on Jobs/CronJobs to affect reconcile loops
		For(&skiperatorv1alpha1.Routing{}).
		Owns(&istionetworkingv1beta1.Gateway{}).
		Owns(&networkingv1.NetworkPolicy{}).
		Owns(&istionetworkingv1beta1.VirtualService{}).
		Watches(&certmanagerv1.Certificate{}, handler.EnqueueRequestsFromMapFunc(r.SkiperatorRoutingCertRequests)).
		WithEventFilter(predicate.Or(predicate.GenerationChangedPredicate{}, predicate.LabelChangedPredicate{})).
		Complete(r)
}

func (r *RoutingReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	routing := &skiperatorv1alpha1.Routing{}
	err := r.GetClient().Get(ctx, req.NamespacedName, routing)

	if errors.IsNotFound(err) {
		return util.DoNotRequeue()
	} else if err != nil {
		r.EmitWarningEvent(routing, "ReconcileStartFail", "something went wrong fetching the SKIPJob, it might have been deleted")
		return util.RequeueWithError(err)
	}

	r.EmitNormalEvent(routing, "ReconcileStart", fmt.Sprintf("Routing %v has started reconciliation loop", routing.Name))

	controllerDuties := []func(context.Context, *skiperatorv1alpha1.Routing) (reconcile.Result, error){
		r.reconcileNetworkPolicy,
		r.reconcileVirtualService,
		r.reconcileGateway,
		r.reconcileCertificate,
	}

	for _, fn := range controllerDuties {
		res, err := fn(ctx, routing)
		if err != nil {
			return res, err
		} else if res.RequeueAfter > 0 || res.Requeue {
			return res, nil
		}
	}

	r.EmitNormalEvent(routing, "ReconcileEnd", fmt.Sprintf("Routing %v has finished reconciliation loop", routing.Name))

	err = r.GetClient().Update(ctx, routing)
	return util.RequeueWithError(err)
}
