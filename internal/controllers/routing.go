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
	"github.com/kartverket/skiperator/pkg/resourcegenerator/resourceutils"
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
	common.ReconcilerBase
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

func (r *RoutingReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	rLog := log.NewLogger().WithName(fmt.Sprintf("routing-controller: %s", req.Name))
	rLog.Debug("Starting reconcile for request", "request", req.Name)

	routing, err := r.getRouting(req, ctx)
	if routing == nil {
		rLog.Info("Routing not found, cleaning up watched resources", "routing", req.Name)
		if errs := r.cleanUpWatchedResources(ctx, req.NamespacedName); len(errs) > 0 {
			return common.RequeueWithError(fmt.Errorf("error when trying to clean up watched resources: %w", errs[0]))
		}
		return common.DoNotRequeue()
	} else if err != nil {
		r.EmitWarningEvent(routing, "ReconcileStartFail", "something went wrong fetching the Routing, it might have been deleted")
		return common.RequeueWithError(err)
	}

	if !common.ShouldReconcile(routing) {
		return common.DoNotRequeue()
	}

	if err := r.setDefaultSpec(routing); err != nil {
		rLog.Error(err, "error when trying to set default spec")
		r.EmitWarningEvent(routing, "ReconcileStartFail", "error when trying to set default spec")
		return common.RequeueWithError(err)
	}

	//Start the actual reconciliation
	rLog.Debug("Starting reconciliation loop", "routing", routing.Name)
	r.SetProgressingState(ctx, routing, fmt.Sprintf("Routing %v has started reconciliation loop", routing.Name))

	istioEnabled := r.IsIstioEnabledForNamespace(ctx, routing.Namespace)
	identityConfigMap, err := r.GetIdentityConfigMap(ctx)
	if err != nil {
		rLog.Error(err, "cant find identity config map")
	}

	reconciliation := NewRoutingReconciliation(ctx, routing, rLog, istioEnabled, r.GetRestConfig(), identityConfigMap)
	resourceGeneration := []reconciliationFunc{
		networkpolicy.Generate,
		virtualservice.Generate,
		gateway.Generate,
		certificate.Generate,
	}

	for _, f := range resourceGeneration {
		if err := f(reconciliation); err != nil {
			rLog.Error(err, "failed to generate routing resource")
			//At this point we don't have the gvk of the resource yet, so we can't set subresource status.
			r.SetErrorState(ctx, routing, err, "failed to generate routing resource", "ResourceGenerationFailure")
			return common.RequeueWithError(err)
		}
	}

	// We need to do this here, so we are sure it's done. Not setting GVK can cause big issues
	if err = r.setRoutingResourceDefaults(reconciliation.GetResources(), routing); err != nil {
		rLog.Error(err, "failed to set routing resource defaults")
		r.SetErrorState(ctx, routing, err, "failed to set routing resource defaults", "ResourceDefaultsFailure")
		return common.RequeueWithError(err)
	}

	if errs := r.GetProcessor().Process(reconciliation); len(errs) > 0 {
		for _, err = range errs {
			rLog.Error(err, "failed to process resource")
			r.EmitWarningEvent(routing, "ReconcileEndFail", fmt.Sprintf("Failed to process routing resources: %s", err.Error()))
		}
		r.SetErrorState(ctx, routing, fmt.Errorf("found %d errors", len(errs)), "failed to process routing resources, see subresource status", "ProcessorFailure")
		return common.RequeueWithError(err)
	}

	r.SetSyncedState(ctx, routing, "Routing has been reconciled")

	return common.DoNotRequeue()
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

func (r *RoutingReconciler) cleanUpWatchedResources(ctx context.Context, name types.NamespacedName) []error {
	route := &skiperatorv1alpha1.Routing{}
	route.SetName(name.Name)
	route.SetNamespace(name.Namespace)

	reconciliation := NewRoutingReconciliation(ctx, route, log.NewLogger(), false, nil, nil)
	return r.GetProcessor().Process(reconciliation)
}

// Do this with application too?
func (r *RoutingReconciler) setDefaultSpec(routing *skiperatorv1alpha1.Routing) error {
	for i := range routing.Spec.Routes {
		route := &routing.Spec.Routes[i] // Get a pointer to the route in the slice
		if route.Port == 0 {
			app, err := r.getTargetApplication(context.Background(), route.TargetApp, routing.Namespace)
			if err != nil {
				return err
			}
			route.Port = int32(app.Spec.Port)
		}
	}
	return nil
}

func (r *RoutingReconciler) setRoutingResourceDefaults(resources []client.Object, routing *skiperatorv1alpha1.Routing) error {
	for _, resource := range resources {
		if err := r.SetSubresourceDefaults(resources, routing); err != nil {
			return err
		}
		resourceutils.SetRoutingLabels(resource, routing)
	}
	return nil
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

// TODO figure out what this does
// TODO have to do something about the hardcoded labels everywhere
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

func (r *RoutingReconciler) getTargetApplication(ctx context.Context, appName string, namespace string) (*skiperatorv1alpha1.Application, error) {
	application := &skiperatorv1alpha1.Application{}
	if err := r.GetClient().Get(ctx, types.NamespacedName{Name: appName, Namespace: namespace}, application); err != nil {
		return nil, fmt.Errorf("error when trying to get target application: %w", err)
	}

	return application, nil
}
