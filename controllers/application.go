package controllers

import (
	"context"

	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/util"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

//+kubebuilder:rbac:groups=skiperator.kartverket.no,resources=applications,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=events,verbs=create;patch
//+kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=serviceaccounts,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=networking.istio.io,resources=gateways,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=networking.k8s.io,resources=networkpolicies,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=networking.istio.io,resources=serviceentries,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=networking.istio.io,resources=virtualservices,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=security.istio.io,resources=peerauthentications,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=autoscaling,resources=horizontalpodautoscalers,verbs=get;list;watch;create;update;patch;delete

type ApplicationReconciler struct {
	util.ReconcilerBase
}

func (r *ApplicationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return newControllerManagedBy[*skiperatorv1alpha1.Application](mgr).
		For(&skiperatorv1alpha1.Application{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&networkingv1beta1.ServiceEntry{}, builder.WithPredicates(
			matchesPredicate[*networkingv1beta1.ServiceEntry](isEgressServiceEntry),
		)).
		Owns(&networkingv1beta1.Gateway{}, builder.WithPredicates(
			matchesPredicate[*networkingv1beta1.Gateway](isIngressGateway),
		)).
		Owns(&autoscalingv2.HorizontalPodAutoscaler{}).
		Owns(&networkingv1beta1.VirtualService{}, builder.WithPredicates(
			matchesPredicate[*networkingv1beta1.VirtualService](isIngressVirtualService),
		)).
		Owns(&securityv1beta1.PeerAuthentication{}).
		Owns(&corev1.ServiceAccount{}).
		WithEventFilter(predicate.GenerationChangedPredicate{}).
		Complete(r)
}

func (r *ApplicationReconciler) Reconcile(ctx context.Context, application *skiperatorv1alpha1.Application) (reconcile.Result, error) {
	_, err := r.initializeApplication(ctx, application)
	if err != nil {
		return reconcile.Result{}, err
	}

	_, err = r.reconcileDeployment(ctx, application)
	if err != nil {
		return reconcile.Result{}, err
	}

	_, err = r.reconcileService(ctx, application)
	if err != nil {
		return reconcile.Result{}, err
	}

	_, err = r.reconcileConfigMap(ctx, application)
	if err != nil {
		return reconcile.Result{}, err
	}

	_, err = r.reconcileEgressServiceEntry(ctx, application)
	if err != nil {
		return reconcile.Result{}, err
	}

	_, err = r.reconcileIngressGateway(ctx, application)
	if err != nil {
		return reconcile.Result{}, err
	}

	_, err = r.reconcileIngressVirtualService(ctx, application)
	if err != nil {
		return reconcile.Result{}, err
	}

	_, err = r.reconcileHorizontalPodAutoscaler(ctx, application)
	if err != nil {
		return reconcile.Result{}, err
	}

	_, err = r.reconcilePeerAuthentication(ctx, application)
	if err != nil {
		return reconcile.Result{}, err
	}

	_, err = r.reconcileServiceAccount(ctx, application)
	if err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, err
}

func (r *ApplicationReconciler) initializeApplication(ctx context.Context, application *skiperatorv1alpha1.Application) (reconcile.Result, error) {

	application.FillDefaults()
	err := r.GetClient().Update(ctx, application)
	if err != nil {
		r.GetRecorder().Eventf(
			application,
			corev1.EventTypeWarning, "Error",
			"Application could not initialize: "+err.Error(),
		)
		return reconcile.Result{}, err
	}

	// This FillDefaults should not have to be called, but the previous application update removes the default statuses.
	// TODO Figure out a way to avoid this second FillDefaults
	application.FillDefaults()
	err = r.GetClient().Status().Update(ctx, application)
	if err != nil {
		r.GetRecorder().Eventf(
			application,
			corev1.EventTypeWarning, "Error",
			"Application could not update status: "+err.Error(),
		)
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, err
}
