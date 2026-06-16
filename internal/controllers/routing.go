package controllers

import (
	"context"
	goerrors "errors"
	"fmt"
	"time"

	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	commontypes "github.com/kartverket/skiperator/api/common"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/internal/controllers/common"
	"github.com/kartverket/skiperator/pkg/gwapi"
	"github.com/kartverket/skiperator/pkg/log"
	"github.com/kartverket/skiperator/pkg/reconciliation"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/certificate"
	gatewayapigenerator "github.com/kartverket/skiperator/pkg/resourcegenerator/gatewayapi"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/istio/gateway"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/istio/virtualservice"
	networkpolicy "github.com/kartverket/skiperator/pkg/resourcegenerator/networkpolicy/dynamic"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/resourceutils"
	"github.com/kartverket/skiperator/pkg/resourceprocessor"
	"github.com/kartverket/skiperator/pkg/resourceschemas"
	istionetworkingv1 "istio.io/client-go/pkg/apis/networking/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	gatewayapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

// sharedRoutingFinalizer drives ref-counted cleanup of shared Gateway API
// resources in istio-gateways, which are not owned by any single contributor.
const sharedRoutingFinalizer = "skiperator.kartverket.no/shared-routing-cleanup"

// +kubebuilder:rbac:groups=skiperator.kartverket.no,resources=routings;routings/status,verbs=get;list;watch;update
// +kubebuilder:rbac:groups=skiperator.kartverket.no,resources=applications;applications/status,verbs=get;list;watch
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch
// +kubebuilder:rbac:groups=networking.k8s.io,resources=networkpolicies,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=networking.istio.io,resources=gateways;virtualservices,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cert-manager.io,resources=certificates,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=gateway.networking.k8s.io,resources=listenersets;httproutes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=gateway.networking.k8s.io,resources=gateways,verbs=get;list;watch

type RoutingReconciler struct {
	common.ReconcilerBase
}

func (r *RoutingReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&skiperatorv1alpha1.Routing{}).
		Owns(&istionetworkingv1.Gateway{}).
		Owns(&networkingv1.NetworkPolicy{}).
		Owns(&istionetworkingv1.VirtualService{}).
		Owns(&gatewayapiv1.ListenerSet{}).
		Owns(&gatewayapiv1.HTTPRoute{}).
		Watches(&certmanagerv1.Certificate{}, handler.EnqueueRequestsFromMapFunc(r.skiperatorRoutingCertRequests)).
		Watches(
			&skiperatorv1alpha1.Application{},
			handler.EnqueueRequestsFromMapFunc(r.skiperatorApplicationsChanges)).
		WithEventFilter(predicate.Or(
			predicate.GenerationChangedPredicate{},
			predicate.LabelChangedPredicate{},
			// Finalizer-driven deletes arrive as metadata-only updates (deletion
			// timestamp set, generation unchanged), so let those through too.
			predicate.Funcs{UpdateFunc: func(e event.UpdateEvent) bool {
				return !e.ObjectNew.GetDeletionTimestamp().IsZero()
			}},
		)).
		Complete(r)
}

func (r *RoutingReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	rLog := log.NewLogger().WithName(fmt.Sprintf("routing-controller: %s", req.Name))
	rLog.Debug("Starting reconcile for request", "request", req.Name)

	routing, err := r.getRouting(ctx, req)
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

	// Shared routing resources in istio-gateways are not owned by any single
	// contributor, so a finalizer drives ref-counted cleanup when the last
	// contributor for a hostname is deleted.
	if handled, result, err := r.reconcileSharedRoutingFinalizer(ctx, routing, rLog); handled {
		return result, err
	}

	if !common.ShouldReconcile(routing) {
		return common.DoNotRequeue()
	}

	tmpStatus := routing.GetStatus().DeepCopy()

	routing.SetDefaultStatus()
	statusDiff, err := common.GetObjectDiff(tmpStatus, routing.GetStatus())
	if err != nil {
		return common.RequeueWithError(err)
	}

	if len(statusDiff) > 0 {
		rLog.Info("Status has changed", "diff", statusDiff)
		routing.GetStatus().SortConditions()
		err = r.GetClient().Status().Update(ctx, routing)
		return reconcile.Result{Requeue: true}, err
	}

	if err := r.setDefaultSpec(ctx, routing); err != nil {
		rLog.Error(err, "error when trying to set default spec")
		r.SetErrorState(ctx, routing, err, "error when trying to set default spec", "DefaultSpecFailure")
		return common.RequeueWithError(err)
	}

	//Start the actual reconciliation
	rLog.Debug("Starting reconciliation loop", "routing", routing.Name)
	r.SetProgressingState(ctx, routing, fmt.Sprintf("Routing %v has started reconciliation loop", routing.Name))

	// Resolve Istio enablement once and reuse it for both the Gateway API
	// prerequisite check and the reconciliation, instead of looking the
	// namespace up twice. A lookup error requeues rather than being read as
	// "Istio disabled".
	istioEnabled, err := r.IsIstioEnabledForNamespace(ctx, routing.Namespace)
	if err != nil {
		rLog.Error(err, "failed to check Istio revision label for namespace")
		r.SetErrorState(ctx, routing, err, "failed to check Istio revision label for namespace", "NamespaceLookupFailure")
		return common.RequeueWithError(err)
	}

	// Gateway API uses shared cluster resources, so fail before generating
	// resources if namespace setup or ownership checks are invalid.
	if checkGatewayAPIPrerequisites(ctx, &r.ReconcilerBase, routing, istioEnabled, rLog) {
		return common.DoNotRequeue()
	}

	reconciliationRouting := reconciliation.NewRoutingReconciliation(ctx, routing, rLog, istioEnabled, r.GetRestConfig())
	routingState, err := gwapi.EvaluateRoutingState(ctx, r.GetClient(), routing, routing.GetStatus())
	if err != nil {
		// A failed routing-state lookup must not be read as "legacy absent":
		// requeue without generating resources so legacy routing is preserved.
		rLog.Error(err, "failed to evaluate Gateway API routing state")
		r.SetErrorState(ctx, routing, err, "failed to evaluate Gateway API routing state", "RoutingStateFailure")
		return common.RequeueWithError(err)
	}
	reconciliationRouting.SetGenerateLegacyRouting(routingState.GenerateLegacyRouting)

	// Prime migration status (start time + stall detection) before generating
	// resources, so the migration clock advances and stalls are surfaced even if
	// resource generation keeps failing. Persisted by the success or error path.
	if routing.UsesStandardRouting() && !routingState.Readiness.Ready {
		emitMigrationEvents(&r.ReconcilerBase, routing, gwapi.UpdateRoutingStatus(routing.GetStatus(), routing.GetGeneration(), routingState))
	}
	resourceGeneration := []reconciliationFunc{
		networkpolicy.Generate,
		virtualservice.Generate,
		gateway.Generate,
		gatewayapigenerator.Generate,
		certificate.Generate,
	}

	for _, f := range resourceGeneration {
		if err := f(reconciliationRouting); err != nil {
			//At this point we don't have the gvk of the resource yet, so we can't set subresource status.
			var subErr *reconciliation.SubResourceError
			if goerrors.As(err, &subErr) {
				rLog.Error(subErr.GetWrapErr(), subErr.Message)
				r.SetErrorState(ctx, routing, subErr.GetWrapErr(), subErr.Message, subErr.GetReason())
			} else {
				// Safe fallback if the error is not of type SubResourceError, to avoid losing error context
				rLog.Error(err, "failed to generate routing resource")
				r.SetErrorState(ctx, routing, err, "failed to generate routing resource", "ResourceGenerationFailure")
			}
			return common.RequeueWithError(err)
		}
	}

	// We need to do this here, so we are sure it's done. Not setting GVK can cause big issues
	if err = r.setRoutingResourceDefaults(reconciliationRouting.GetResources(), routing); err != nil {
		rLog.Error(err, "failed to set routing resource defaults")
		r.SetErrorState(ctx, routing, err, "failed to set routing resource defaults", "ResourceDefaultsFailure")
		return common.RequeueWithError(err)
	}

	processor := resourceprocessor.NewResourceProcessor(r.GetClient(), resourceschemas.GetRoutingSchemas(r.GetScheme()), r.GetScheme())

	if errs := processor.Process(reconciliationRouting); len(errs) > 0 {
		for _, err = range errs {
			rLog.Error(err, "failed to process resource")
			r.EmitWarningEvent(routing, "ReconcileEndFail", fmt.Sprintf("Failed to process routing resources: %s", err.Error()))
		}
		r.SetErrorState(ctx, routing, fmt.Errorf("found %d errors", len(errs)), "failed to process routing resources, see subresource status", "ProcessorFailure")
		return common.RequeueWithError(err)
	}

	// Register as a contributor to the shared hostname so shared istio-gateways
	// resources are ref-counted for garbage collection. The matching deregister
	// runs in the finalizer.
	if routing.UsesSharedOwnership() {
		host, err := routing.Spec.GetHost()
		if err != nil {
			r.SetErrorState(ctx, routing, err, "failed to resolve shared routing host", "SharedRoutingHostFailure")
			return common.RequeueWithError(err)
		}
		if err := gwapi.RegisterSharedContributor(ctx, r.GetClient(), host.Hostname, types.NamespacedName{Namespace: routing.Namespace, Name: routing.Name}); err != nil {
			r.SetErrorState(ctx, routing, err, "failed to register shared routing contributor", "SharedRoutingMembershipFailure")
			return common.RequeueWithError(err)
		}
	}

	// Ready/summary come from the shared routing-status assembler, the same path
	// Application uses, so the two controllers cannot drift.
	finalizeRoutingStatus(&r.ReconcilerBase, routing, routingState, "Routing has been reconciled")
	if routing.UsesStandardRouting() {
		setSharedRoutingResourcesCondition(routing)
	}
	r.UpdateStatus(ctx, routing)
	if routing.UsesStandardRouting() && !routingState.Readiness.Ready {
		return reconcile.Result{RequeueAfter: 10 * time.Second}, nil
	}

	return common.DoNotRequeue()
}

// setSharedRoutingResourcesCondition exposes shared Gateway API infrastructure
// only when active, without adding conditions to standalone Routing objects.
func setSharedRoutingResourcesCondition(routing *skiperatorv1alpha1.Routing) {
	if routing.UsesSharedOwnership() {
		routing.GetStatus().SetSharedRoutingResourcesCondition(
			metav1.ConditionTrue,
			routing.GetGeneration(),
			"SharedRoutingResourcesActive",
			"Routing uses shared Gateway API resources in istio-gateways",
		)
		return
	}
	meta.RemoveStatusCondition(&routing.GetStatus().Conditions, commontypes.SharedRoutingResourcesType)
}

func (r *RoutingReconciler) getRouting(ctx context.Context, req reconcile.Request) (*skiperatorv1alpha1.Routing, error) {
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

	reconciliation := reconciliation.NewRoutingReconciliation(ctx, route, log.NewLogger(), false, nil)

	processor := resourceprocessor.NewResourceProcessor(r.GetClient(), resourceschemas.GetRoutingSchemas(r.GetScheme()), r.GetScheme())
	return processor.Process(reconciliation)
}

// reconcileSharedRoutingFinalizer adds/removes the shared-routing finalizer and,
// on deletion, performs ref-counted cleanup of shared istio-gateways resources.
// It returns handled=true when it owns the outcome of this reconcile pass.
func (r *RoutingReconciler) reconcileSharedRoutingFinalizer(ctx context.Context, routing *skiperatorv1alpha1.Routing, rLog log.Logger) (bool, reconcile.Result, error) {
	if !routing.GetDeletionTimestamp().IsZero() {
		if !ctrlutil.ContainsFinalizer(routing, sharedRoutingFinalizer) {
			return true, reconcile.Result{}, nil
		}
		if err := r.handleSharedRoutingDeletion(ctx, routing, rLog); err != nil {
			return true, reconcile.Result{}, err
		}
		ctrlutil.RemoveFinalizer(routing, sharedRoutingFinalizer)
		if err := r.GetClient().Update(ctx, routing); err != nil {
			return true, reconcile.Result{}, err
		}
		return true, reconcile.Result{}, nil
	}

	hasFinalizer := ctrlutil.ContainsFinalizer(routing, sharedRoutingFinalizer)
	// Add the finalizer for shared routings, remove it if a Routing switched away
	// from shared ownership. Requeue so the rest of the reconcile runs cleanly.
	if routing.UsesSharedOwnership() && !hasFinalizer {
		ctrlutil.AddFinalizer(routing, sharedRoutingFinalizer)
		if err := r.GetClient().Update(ctx, routing); err != nil {
			return true, reconcile.Result{}, err
		}
		return true, reconcile.Result{Requeue: true}, nil
	}
	if !routing.UsesSharedOwnership() && hasFinalizer {
		// Switched away from shared ownership: release membership so the shared
		// resources are cleaned up if this was the last contributor.
		if err := r.releaseSharedMembership(ctx, routing, rLog); err != nil {
			return true, reconcile.Result{}, err
		}
		ctrlutil.RemoveFinalizer(routing, sharedRoutingFinalizer)
		if err := r.GetClient().Update(ctx, routing); err != nil {
			return true, reconcile.Result{}, err
		}
		return true, reconcile.Result{Requeue: true}, nil
	}
	return false, reconcile.Result{}, nil
}

// handleSharedRoutingDeletion prunes the deleted Routing's own resources, then
// releases its shared-routing membership.
func (r *RoutingReconciler) handleSharedRoutingDeletion(ctx context.Context, routing *skiperatorv1alpha1.Routing, rLog log.Logger) error {
	if errs := r.cleanUpWatchedResources(ctx, types.NamespacedName{Namespace: routing.Namespace, Name: routing.Name}); len(errs) > 0 {
		return fmt.Errorf("failed to clean up routing resources: %w", errs[0])
	}
	return r.releaseSharedMembership(ctx, routing, rLog)
}

// releaseSharedMembership deregisters this Routing from its hostname's shared
// membership and, when it was the last contributor, deletes the shared
// istio-gateways resources and the membership ConfigMap. Used both on deletion
// and when a Routing switches away from shared ownership.
func (r *RoutingReconciler) releaseSharedMembership(ctx context.Context, routing *skiperatorv1alpha1.Routing, rLog log.Logger) error {
	host, err := routing.Spec.GetHost()
	if err != nil {
		return err
	}
	empty, err := gwapi.DeregisterSharedContributor(ctx, r.GetClient(), host.Hostname, types.NamespacedName{Namespace: routing.Namespace, Name: routing.Name})
	if err != nil {
		return err
	}
	if !empty {
		rLog.Debug("Other shared Routings still use hostname, keeping shared resources", "hostname", host.Hostname)
		return nil
	}
	if err := r.deleteSharedRoutingResources(ctx, routing, host); err != nil {
		return err
	}
	return gwapi.DeleteSharedMembership(ctx, r.GetClient(), host.Hostname)
}

// deleteSharedRoutingResources deletes the shared ListenerSet, redirect HTTPRoute
// and certificate for a hostname from istio-gateways. Missing resources are
// ignored so the cleanup is idempotent.
func (r *RoutingReconciler) deleteSharedRoutingResources(ctx context.Context, routing *skiperatorv1alpha1.Routing, host *commontypes.Host) error {
	resources := []client.Object{
		&gatewayapiv1.ListenerSet{ObjectMeta: metav1.ObjectMeta{Namespace: gwapi.IstioGatewayNamespace, Name: gwapi.SharedListenerSetName(host.Hostname)}},
		&gatewayapiv1.HTTPRoute{ObjectMeta: metav1.ObjectMeta{Namespace: gwapi.IstioGatewayNamespace, Name: gwapi.SharedRedirectRouteName(host.Hostname)}},
	}
	if !host.UsesCustomCert() {
		certName, err := routing.GetCertificateName(host)
		if err != nil {
			return err
		}
		resources = append(resources, &certmanagerv1.Certificate{ObjectMeta: metav1.ObjectMeta{Namespace: gwapi.IstioGatewayNamespace, Name: certName}})
	}

	for _, obj := range resources {
		if err := r.GetClient().Delete(ctx, obj); err != nil && !errors.IsNotFound(err) {
			return fmt.Errorf("failed to delete shared routing resource %s/%s: %w", obj.GetNamespace(), obj.GetName(), err)
		}
	}
	return nil
}

// TODO Do this with application too for dynamic port allocation?
func (r *RoutingReconciler) setDefaultSpec(ctx context.Context, routing *skiperatorv1alpha1.Routing) error {
	for i := range routing.Spec.Routes {
		route := &routing.Spec.Routes[i] // Get a pointer to the route in the slice
		if route.Port == 0 {
			app, err := r.getTargetApplication(ctx, route.TargetApp, routing.Namespace)
			if err != nil {
				return err
			}
			route.Port = int32(app.Spec.Port)
		}
	}
	return nil
}

func (r *RoutingReconciler) setRoutingResourceDefaults(resources []client.Object, routing *skiperatorv1alpha1.Routing) error {
	host, err := routing.Spec.GetHost()
	if err != nil {
		return err
	}
	for _, resource := range resources {
		if err := r.SetSubresourceDefaults(resources, routing); err != nil {
			return err
		}
		if routing.UsesSharedOwnership() && resource.GetNamespace() == gwapi.IstioGatewayNamespace && isSharedRoutingInfrastructure(resource) {
			resourceutils.SetSharedRoutingLabels(resource, host.Hostname)
			continue
		}
		resourceutils.SetRoutingLabels(resource, routing)
	}
	return nil
}

// isSharedRoutingInfrastructure selects the cross-namespace objects whose
// labels must be stable across all Routing contributors for one hostname.
func isSharedRoutingInfrastructure(resource client.Object) bool {
	switch resource.(type) {
	case *certmanagerv1.Certificate, *gatewayapiv1.ListenerSet, *gatewayapiv1.HTTPRoute:
		return true
	default:
		return false
	}
}

func (r *RoutingReconciler) skiperatorApplicationsChanges(context context.Context, obj client.Object) []reconcile.Request {
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
func (r *RoutingReconciler) skiperatorRoutingCertRequests(_ context.Context, obj client.Object) []reconcile.Request {
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
				Namespace: certificate.Labels["skiperator.kartverket.no/source-namespace"],
				Name:      certificate.Labels["skiperator.kartverket.no/routing-name"],
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
