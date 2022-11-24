package controllers

import (
	"context"
	"time"

	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

//+kubebuilder:rbac:groups=skiperator.kartverket.no,resources=applications,verbs=get;list;watch

type ApplicationReconciler struct {
	client   client.Client
	scheme   *runtime.Scheme
	recorder record.EventRecorder
}

func (r *ApplicationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.client = mgr.GetClient()
	r.scheme = mgr.GetScheme()
	r.recorder = mgr.GetEventRecorderFor("application-controller")

	return ctrl.NewControllerManagedBy(mgr).
		For(&skiperatorv1alpha1.Application{}).
		Owns(&skiperatorv1alpha1.Application{}).
		Complete(r)
}

func (r *ApplicationReconciler) Reconcile(ctx context.Context, req reconcile.Request) (ctrl.Result, error) {
	application := skiperatorv1alpha1.Application{ObjectMeta: metav1.ObjectMeta{Namespace: req.Namespace, Name: req.Name}}

	_, err := ctrlutil.CreateOrPatch(ctx, r.client, &application, func() error {

		application.Status.TotalApplicationStatus = skiperatorv1alpha1.Status{Status: skiperatorv1alpha1.StatusNames("Progressing"), Message: "Starting reconcile loop", TimeStamp: time.Now().String()}
		ControllerStatus := make(map[string]skiperatorv1alpha1.Status)
		ControllerStatus[string("Deployment")] = skiperatorv1alpha1.Status{Status: skiperatorv1alpha1.StatusNames("Progressing"), Message: "Starting reconcile loop", TimeStamp: time.Now().String()}
		application.Status.ControllersApplicationStatus = ControllerStatus
		return nil

	})

	return reconcile.Result{}, err
}
