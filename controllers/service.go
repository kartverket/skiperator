package controllers

import (
	"context"

	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

//+kubebuilder:rbac:groups=skiperator.kartverket.no,resources=applications,verbs=get;list;watch
//+kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete

func (r *ApplicationReconciler) reconcileService(ctx context.Context, application *skiperatorv1alpha1.Application) (reconcile.Result, error) {
	application.FillDefaults()
	controllerName := "service"
	controllerMessageName := "Service"
	r.ManageControllerStatus(ctx, application, controllerName, skiperatorv1alpha1.Status{Message: controllerMessageName + " starting reconciliation", Status: skiperatorv1alpha1.PROGRESSING})

	service := corev1.Service{ObjectMeta: metav1.ObjectMeta{Namespace: application.Namespace, Name: application.Name}}
	_, err := ctrlutil.CreateOrPatch(ctx, r.GetClient(), &service, func() error {
		// Set application as owner of the service
		err := ctrlutil.SetControllerReference(application, &service, r.GetScheme())
		if err != nil {
			r.ManageControllerStatus(ctx, application, controllerName, skiperatorv1alpha1.Status{Message: controllerMessageName + " encountered error: " + err.Error(), Status: skiperatorv1alpha1.ERROR})
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

	if err != nil {
		r.ManageControllerStatus(ctx, application, controllerName, skiperatorv1alpha1.Status{Message: controllerMessageName + " encountered error: " + err.Error(), Status: skiperatorv1alpha1.ERROR})
	} else {
		r.ManageControllerStatus(ctx, application, controllerName, skiperatorv1alpha1.Status{Message: controllerMessageName + " synced", Status: skiperatorv1alpha1.SYNCED})
	}

	return reconcile.Result{}, err
}
