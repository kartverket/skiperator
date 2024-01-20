package applicationcontroller

import (
	"context"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/api/v1alpha1/podtypes"
	"github.com/kartverket/skiperator/pkg/util"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"strings"
)

const defaultPortName = "http"

var defaultPrometheusPort = corev1.ServicePort{
	Name:       util.IstioMetricsPortName.StrVal,
	Protocol:   corev1.ProtocolTCP,
	Port:       util.IstioMetricsPortNumber.IntVal,
	TargetPort: util.IstioMetricsPortNumber,
}

func (r *ApplicationReconciler) reconcileService(ctx context.Context, application *skiperatorv1alpha1.Application) (reconcile.Result, error) {
	controllerName := "Service"
	r.SetControllerProgressing(ctx, application, controllerName)

	service := corev1.Service{ObjectMeta: metav1.ObjectMeta{Namespace: application.Namespace, Name: application.Name}}
	shouldReconcile, err := r.ShouldReconcile(ctx, &service)
	if err != nil || !shouldReconcile {
		r.SetControllerFinishedOutcome(ctx, application, controllerName, err)
		return util.RequeueWithError(err)
	}

	_, err = ctrlutil.CreateOrPatch(ctx, r.GetClient(), &service, func() error {
		// Set application as owner of the service
		err := ctrlutil.SetControllerReference(application, &service, r.GetScheme())
		if err != nil {
			r.SetControllerError(ctx, application, controllerName, err)
			return err
		}

		r.SetLabelsFromApplication(&service, *application)
		util.SetCommonAnnotations(&service)

		// ServiceMonitor requires labels to be set on service to select it
		labels := service.GetLabels()
		if len(labels) == 0 {
			labels = make(map[string]string)
		}
		labels["app"] = application.Name
		service.SetLabels(labels)

		ports := append(getAdditionalPorts(application.Spec.AdditionalPorts), getServicePort(application.Spec.Port, application.Spec.AppProtocol))
		if r.IsIstioEnabledForNamespace(ctx, application.Namespace) {
			ports = append(ports, defaultPrometheusPort)
		}

		service.Spec = corev1.ServiceSpec{
			Selector: util.GetPodAppSelector(application.Name),
			Type:     corev1.ServiceTypeClusterIP,
			Ports:    ports,
		}

		return nil
	})

	r.SetControllerFinishedOutcome(ctx, application, controllerName, err)

	return util.RequeueWithError(err)
}

func getAdditionalPorts(additionalPorts []podtypes.InternalPort) []corev1.ServicePort {
	var ports []corev1.ServicePort

	for _, p := range additionalPorts {
		ports = append(ports, corev1.ServicePort{
			Name:       p.Name,
			Port:       p.Port,
			Protocol:   p.Protocol,
			TargetPort: intstr.FromInt(int(p.Port)),
		})
	}

	return ports
}

func getServicePort(port int, appProtocol string) corev1.ServicePort {
	var resolvedProtocol = corev1.ProtocolTCP
	if strings.ToLower(appProtocol) == "udp" {
		resolvedProtocol = corev1.ProtocolUDP
	}

	var resolvedAppProtocol = appProtocol
	if len(resolvedAppProtocol) == 0 {
		resolvedAppProtocol = "http"
	} else if port == 5432 {
		// Legacy postgres hack
		resolvedAppProtocol = "tcp"
	}

	return corev1.ServicePort{
		Name:        defaultPortName,
		Protocol:    resolvedProtocol,
		AppProtocol: &resolvedAppProtocol,
		Port:        int32(port),
		TargetPort:  intstr.FromInt(port),
	}
}
