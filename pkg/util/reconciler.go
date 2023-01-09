package util

import (
	"context"
	"strings"

	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"golang.org/x/exp/maps"
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

func (r *ReconcilerBase) manageControllerStatus(context context.Context, app *skiperatorv1alpha1.Application, controller string, statusName skiperatorv1alpha1.StatusNames, message string) (reconcile.Result, error) {
	app.UpdateControllerStatus(controller, message, statusName)
	err := r.GetClient().Status().Update(context, app)
	if err != nil {
		return reconcile.Result{}, err
	}
	return reconcile.Result{}, nil
}

func (r *ReconcilerBase) manageControllerStatusError(context context.Context, app *skiperatorv1alpha1.Application, controller string, issue error) (reconcile.Result, error) {
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

func (r *ReconcilerBase) SetControllerPending(context context.Context, app *skiperatorv1alpha1.Application, controller string) (reconcile.Result, error) {
	message := controller + " has been initialized and is pending Skiperator startup"

	return r.manageControllerStatus(context, app, controller, skiperatorv1alpha1.PENDING, message)
}

func (r *ReconcilerBase) SetControllerProgressing(context context.Context, app *skiperatorv1alpha1.Application, controller string) (reconcile.Result, error) {
	message := controller + " has started sync"

	return r.manageControllerStatus(context, app, controller, skiperatorv1alpha1.PROGRESSING, message)
}

func (r *ReconcilerBase) SetControllerSynced(context context.Context, app *skiperatorv1alpha1.Application, controller string) (reconcile.Result, error) {
	message := controller + " has finished synchronizing"

	return r.manageControllerStatus(context, app, controller, skiperatorv1alpha1.SYNCED, message)
}

func (r *ReconcilerBase) SetControllerError(context context.Context, app *skiperatorv1alpha1.Application, controller string, issue error) (reconcile.Result, error) {
	return r.manageControllerStatusError(context, app, controller, issue)
}

func (r *ReconcilerBase) SetControllerFinishedOutcome(context context.Context, app *skiperatorv1alpha1.Application, controllerName string, issue error) (reconcile.Result, error) {
	if issue != nil {
		return r.manageControllerStatusError(context, app, controllerName, issue)
	}

	return r.SetControllerSynced(context, app, controllerName)
}

func (r *ReconcilerBase) setResourceLabelsIfApplies(context context.Context, obj client.Object, app skiperatorv1alpha1.Application) {
	objectGroupVersionKind := obj.GetObjectKind().GroupVersionKind()

	for controllerResource, resourceLabels := range app.Spec.ResourceLabels {
		resourceLabelGroupKind, present := app.GroupKindFromControllerResource(controllerResource)
		if present {
			if strings.EqualFold(objectGroupVersionKind.Group, resourceLabelGroupKind.Group) && strings.EqualFold(objectGroupVersionKind.Kind, resourceLabelGroupKind.Kind) {
				objectLabels := obj.GetLabels()
				maps.Copy(objectLabels, resourceLabels)
				obj.SetLabels(objectLabels)
			}
		} else {
			r.GetRecorder().Eventf(
				&app,
				corev1.EventTypeWarning, "MistypedLabel",
				"Could not find according Kind for Resource "+controllerResource+". Make sure your resource is spelled correctly",
			)
		}

	}
}

func (r *ReconcilerBase) SetLabelsFromApplication(context context.Context, object client.Object, app skiperatorv1alpha1.Application) {
	labels := object.GetLabels()
	if len(labels) == 0 {
		labels = make(map[string]string)
	}
	maps.Copy(labels, app.Spec.Labels)
	object.SetLabels(labels)

	r.setResourceLabelsIfApplies(context, object, app)
}
