package controllers

import (
	"context"
	"fmt"
	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/log"
	"github.com/kartverket/skiperator/pkg/util"
	istionetworkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
		For(&skiperatorv1alpha1.Routing{}).
		Owns(&istionetworkingv1beta1.Gateway{}).
		Owns(&networkingv1.NetworkPolicy{}).
		Owns(&istionetworkingv1beta1.VirtualService{}).
		Watches(&certmanagerv1.Certificate{}, handler.EnqueueRequestsFromMapFunc(r.SkiperatorRoutingCertRequests)).
		Watches(
			&skiperatorv1alpha1.Application{},
			handler.EnqueueRequestsFromMapFunc(r.SkiperatorApplicationsChanges)).
		WithEventFilter(predicate.Or(predicate.GenerationChangedPredicate{}, predicate.LabelChangedPredicate{})).
		Complete(r)
}

func (r *RoutingReconciler) getRouting(req reconcile.Request, ctx context.Context) (*skiperatorv1alpha1.Routing, error) {
	ctxLog := log.FromContext(ctx)
	ctxLog.Debug("Trying to get routing from request", req)

	routing := &skiperatorv1alpha1.Routing{}
	if err := r.GetClient().Get(ctx, req.NamespacedName, routing); err != nil {
		if errors.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("error when trying to get routing: %w", err)
	}

	return routing, nil
}

func (r *RoutingReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	ctxLog := log.NewLogger(ctx).WithName("routing-controller")
	ctxLog.Debug("Starting reconcile for request", req.Name)

	routing, err := r.getRouting(req, ctx)
	if routing == nil {
		return util.DoNotRequeue()
	} else if err != nil {
		r.EmitWarningEvent(routing, "ReconcileStartFail", "something went wrong fetching the Routing, it might have been deleted")
		return util.RequeueWithError(err)
	}

	//Start the actual reconciliation
	ctxLog.Debug("Starting reconciliation loop", routing.Name)
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

func (r *RoutingReconciler) SkiperatorApplicationsChanges(context context.Context, obj client.Object) []reconcile.Request {
	application, isApplication := obj.(*skiperatorv1alpha1.Application)

	if !isApplication {
		return nil
	}

	// List all routings in the same namespace as the application
	routesList := &skiperatorv1alpha1.RoutingList{}
	if err := r.GetClient().List(context, routesList, &client.ListOptions{Namespace: application.Namespace}); err != nil {
		return nil
	}

	// Create a list of reconcile.Requests for each Routing in the same namespace as the application
	requests := make([]reconcile.Request, 0)
	for _, route := range routesList.Items {
		requests = append(requests, reconcile.Request{
			NamespacedName: types.NamespacedName{
				Namespace: route.Namespace,
				Name:      route.Name,
			},
		})
	}

	return requests
}
