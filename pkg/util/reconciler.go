package util

import (
	"context"

	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// ReconcilerBase is a base struct from which all reconcilers can be derived from. By doing so your reconcilers will also inherit a set of utility functions
// To inherit from reconciler just build your finalizer this way:
//
//	type MyReconciler struct {
//	  util.ReconcilerBase
//	  ... other optional fields ...
//	}
type ReconcilerBase struct {
	apireader  client.Reader
	client     client.Client
	scheme     *runtime.Scheme
	restConfig *rest.Config
	recorder   record.EventRecorder
}

func NewReconcilerBase(client client.Client, scheme *runtime.Scheme, restConfig *rest.Config, recorder record.EventRecorder, apireader client.Reader) ReconcilerBase {
	return ReconcilerBase{
		apireader:  apireader,
		client:     client,
		scheme:     scheme,
		restConfig: restConfig,
		recorder:   recorder,
	}
}

// NewReconcilerBase is a contruction function to create a new ReconcilerBase.
func NewFromManager(mgr manager.Manager, recorder record.EventRecorder) ReconcilerBase {
	return NewReconcilerBase(mgr.GetClient(), mgr.GetScheme(), mgr.GetConfig(), recorder, mgr.GetAPIReader())
}

// GetClient returns the underlying client
func (r *ReconcilerBase) GetClient() client.Client {
	return r.client
}

// GetRestConfig returns the undelying rest config
func (r *ReconcilerBase) GetRestConfig() *rest.Config {
	return r.restConfig
}

// GetRecorder returns the underlying recorder
func (r *ReconcilerBase) GetRecorder() record.EventRecorder {
	return r.recorder
}

// GetScheme returns the scheme
func (r *ReconcilerBase) GetScheme() *runtime.Scheme {
	return r.scheme
}

func (r *ReconcilerBase) ManageControllerStatus(context context.Context, app *skiperatorv1alpha1.Application, controller string, statusName skiperatorv1alpha1.StatusNames) (reconcile.Result, error) {
	message := ""
	switch statusName {
	case skiperatorv1alpha1.PROGRESSING:
		message = controller + " has started sync"
	case skiperatorv1alpha1.SYNCED:
		message = controller + " has finished synchronizing"
	case skiperatorv1alpha1.PENDING:
		message = controller + " has been initialized and is pending Skiperator startup"
	default:
		message = "Unknown StatusName in Controller status, not sure how to respond"
	}
	app.UpdateControllerStatus(controller, message, statusName)
	err := r.GetClient().Status().Update(context, app)
	if err != nil {
		return reconcile.Result{}, err
	}
	return reconcile.Result{}, nil
}

func (r *ReconcilerBase) ManageControllerStatusError(context context.Context, app *skiperatorv1alpha1.Application, controller string, issue error) (reconcile.Result, error) {
	app.UpdateControllerStatus(controller, issue.Error(), skiperatorv1alpha1.ERROR)
	err := r.GetClient().Status().Update(context, app)
	r.GetRecorder().Eventf(
		app,
		corev1.EventTypeWarning, "Controller Fault",
		controller+" controller experienced an error",
	)
	if err != nil {
		return reconcile.Result{}, err
	}
	return reconcile.Result{}, issue
}

func (r *ReconcilerBase) ManageControllerOutcome(context context.Context, app *skiperatorv1alpha1.Application, controllerName string, statusName skiperatorv1alpha1.StatusNames, issue error) (reconcile.Result, error) {
	if issue != nil {
		return r.ManageControllerStatusError(context, app, controllerName, issue)
	}

	return r.ManageControllerStatus(context, app, controllerName, statusName)
}
