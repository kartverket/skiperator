package controllers

import (
	"context"

	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

//+kubebuilder:rbac:groups=skiperator.kartverket.no,resources=applications,verbs=get;list;watch
//+kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete

type ServiceReconciler struct {
	client client.Client
	scheme *runtime.Scheme
}

func (r *ServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.client = mgr.GetClient()
	r.scheme = mgr.GetScheme()

	return newControllerManagedBy[*skiperatorv1alpha1.Application](mgr).
		For(&skiperatorv1alpha1.Application{}).
		Owns(&corev1.Service{}).
		Complete(r)
}

func (r *ServiceReconciler) Reconcile(ctx context.Context, application *skiperatorv1alpha1.Application) (reconcile.Result, error) {
	application.FillDefaults()

	application.UpdateControllerStatus("service", "Starting service reconciliation", skiperatorv1alpha1.PROGRESSING)
	r.client.Status().Update(ctx, application)

	service := corev1.Service{ObjectMeta: metav1.ObjectMeta{Namespace: application.Namespace, Name: application.Name}}
	_, err := ctrlutil.CreateOrPatch(ctx, r.client, &service, func() error {
		// Set application as owner of the service
		err := ctrlutil.SetControllerReference(application, &service, r.scheme)
		if err != nil {
			return err
		}

		labels := map[string]string{"app": application.Name}
		service.Spec.Selector = labels

		service.Spec.Type = corev1.ServiceTypeClusterIP

		service.Spec.Ports = make([]corev1.ServicePort, 1)
		service.Spec.Ports[0].Port = int32(application.Spec.Port)
		service.Spec.Ports[0].TargetPort = intstr.FromInt(application.Spec.Port)
		if application.Spec.Port == 5432 { // TODO: Should not be hardcoded
			tcp := "tcp"
			service.Spec.Ports[0].Name = "tcp"
			service.Spec.Ports[0].AppProtocol = &tcp
		} else {
			http := "http"
			service.Spec.Ports[0].Name = "http"
			service.Spec.Ports[0].AppProtocol = &http
		}

		return nil
	})

	application.UpdateControllerStatus("service", "Service synced", skiperatorv1alpha1.SYNCED)
	r.client.Status().Update(ctx, application)

	return reconcile.Result{}, err
}
