package applicationcontroller

import (
	"context"

	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/util"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *ApplicationReconciler) reconcileService(ctx context.Context, application *skiperatorv1alpha1.Application) (reconcile.Result, error) {
	controllerName := "Service"
	r.SetControllerProgressing(ctx, application, controllerName)

	service := corev1.Service{ObjectMeta: metav1.ObjectMeta{Namespace: application.Namespace, Name: application.Name}}
	_, err := ctrlutil.CreateOrPatch(ctx, r.GetClient(), &service, func() error {
		// Set application as owner of the service
		err := ctrlutil.SetControllerReference(application, &service, r.GetScheme())
		if err != nil {
			r.SetControllerError(ctx, application, controllerName, err)
			return err
		}

		r.SetLabelsFromApplication(ctx, &service, *application)
		util.SetCommonAnnotations(&service)

		service.Spec = corev1.ServiceSpec{
			Selector: util.GetPodAppSelector(application.Name),
			Type:     corev1.ServiceTypeClusterIP,
			Ports: []corev1.ServicePort{
				getServicePort(application.Spec.Port),
			},
		}

		return nil
	})

	r.SetControllerFinishedOutcome(ctx, application, controllerName, err)

	return reconcile.Result{}, err
}

func getServicePort(applicationPort int) corev1.ServicePort {
	nameAndProtocol := "http"

	// TODO: Should not be hardcoded
	if applicationPort == 5432 {
		nameAndProtocol = "tcp"
	}

	return corev1.ServicePort{
		Name:        nameAndProtocol,
		AppProtocol: &nameAndProtocol,
		Port:        int32(applicationPort),
		TargetPort:  intstr.FromInt(applicationPort),
	}
}
