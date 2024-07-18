package controllers

import (
	"context"
	"fmt"
	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/internal/controllers/common"
	"github.com/kartverket/skiperator/pkg/log"
	. "github.com/kartverket/skiperator/pkg/reconciliation"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/certificate"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/istio/gateway"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/istio/virtualservice"
	networkpolicy "github.com/kartverket/skiperator/pkg/resourcegenerator/networkpolicy/dynamic"
	istionetworkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// +kubebuilder:rbac:groups=skiperator.kartverket.no,resources=routings;routings/status,verbs=get;list;watch;update
// +kubebuilder:rbac:groups=skiperator.kartverket.no,resources=applications;applications/status,verbs=get;list;watch
// +kubebuilder:rbac:groups=networking.k8s.io,resources=networkpolicies,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=networking.istio.io,resources=gateways;virtualservices,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cert-manager.io,resources=certificates,verbs=get;list;watch;create;update;patch;delete

type RoutingReconciler struct {
	common.ReconcilerBase
}

// TODO fix this
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
		//TODO why red?
		Complete(r)
}

func (r *RoutingReconciler) getRouting(req reconcile.Request, ctx context.Context) (*skiperatorv1alpha1.Routing, error) {
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
	rLog := log.NewLogger().WithName(fmt.Sprintf("routing: %s", req.Name))
	rLog.Debug("Starting reconcile for request", "request", req.Name)

	routing, err := r.getRouting(req, ctx)
	if routing == nil {
		return common.DoNotRequeue()
	} else if err != nil {
		r.EmitWarningEvent(routing, "ReconcileStartFail", "something went wrong fetching the Routing, it might have been deleted")
		return common.RequeueWithError(err)
	}

	identityConfigMap, err := r.GetIdentityConfigMap(ctx)
	if err != nil {
		rLog.Error(err, "cant find identity config map")
	}

	//Start the actual reconciliation
	rLog.Debug("Starting reconciliation loop", "routing", routing.Name)
	r.EmitNormalEvent(routing, "ReconcileStart", fmt.Sprintf("Routing %v has started reconciliation loop", routing.Name))

	reconciliation := NewRoutingReconciliation(ctx, routing, rLog, r.GetRestConfig(), identityConfigMap)
	resourceGeneration := []reconciliationFunc{
		networkpolicy.Generate,
		virtualservice.Generate,
		gateway.Generate,
		certificate.Generate,
	}

	for _, f := range resourceGeneration {
		if err := f(reconciliation); err != nil {
			return common.RequeueWithError(err)
		}
	}

	if err = r.GetProcessor().Process(reconciliation); err != nil {
		return common.RequeueWithError(err)
	}

	r.GetClient().Status().Update(ctx, routing)

	r.EmitNormalEvent(routing, "ReconcileEnd", fmt.Sprintf("Routing %v has finished reconciliation loop", "routing", routing.Name))

	return common.DoNotRequeue()
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

// TODO Set labels
func getRoutingLabels(obj client.Object, routing *skiperatorv1alpha1.Routing) map[string]string {
	labels := make(map[string]string)

	labels["app.kubernetes.io/managed-by"] = "skiperator"
	labels["skiperator.kartverket.no/controller"] = "routing"
	labels["skiperator.kartverket.no/source-namespace"] = routing.Namespace

	return labels
}

// TODO figure out what this does
func (r *RoutingReconciler) SkiperatorRoutingCertRequests(_ context.Context, obj client.Object) []reconcile.Request {
	certificate, isCert := obj.(*certmanagerv1.Certificate)

	if !isCert {
		return nil
	}

	isSkiperatorRoutingOwned := certificate.Labels["app.kubernetes.io/managed-by"] == "skiperator" &&
		certificate.Labels["skiperator.kartverket.no/controller"] == "routing"

	requests := make([]reconcile.Request, 0)

	if isSkiperatorRoutingOwned {
		requests = append(requests, reconcile.Request{
			NamespacedName: types.NamespacedName{
				Namespace: certificate.Labels["application.skiperator.no/app-namespace"],
				Name:      certificate.Labels["application.skiperator.no/app-name"],
			},
		})
	}

	return requests
}
