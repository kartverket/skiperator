package service

import (
	"fmt"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/api/v1alpha1/podtypes"
	"github.com/kartverket/skiperator/pkg/reconciliation"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/resourceutils"
	"github.com/kartverket/skiperator/pkg/util"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

const defaultPortName = "http"

var defaultPrometheusPort = corev1.ServicePort{
	Name:       util.IstioMetricsPortName.StrVal,
	Protocol:   corev1.ProtocolTCP,
	Port:       util.IstioMetricsPortNumber.IntVal,
	TargetPort: util.IstioMetricsPortNumber,
}

func Generate(r reconciliation.Reconciliation) error {
	ctxLog := r.GetLogger()
	if r.GetType() != reconciliation.ApplicationType {
		return fmt.Errorf("unsupported type %s in deployment resource", r.GetType())
	}
	application, ok := r.GetSKIPObject().(*skiperatorv1alpha1.Application)
	if !ok {
		err := fmt.Errorf("failed to cast resource to application")
		ctxLog.Error(err, "Failed to generate deployment resource")
		return err
	}
	ctxLog.Debug("Attempting to create service for application", "application", application.Name)

	service := corev1.Service{ObjectMeta: metav1.ObjectMeta{Namespace: application.Namespace, Name: application.Name}}
	resourceutils.SetApplicationLabels(&service, application)

	ports := append(getAdditionalPorts(application.Spec.AdditionalPorts), getServicePort(application.Spec.Port, application.Spec.AppProtocol))
	if r.IsIstioEnabled() {
		ports = append(ports, defaultPrometheusPort)
	}

	service.Spec = corev1.ServiceSpec{
		Selector: util.GetPodAppSelector(application.Name),
		Type:     corev1.ServiceTypeClusterIP,
		Ports:    ports,
	}

	ctxLog.Debug("created service manifest for application", "application", application.Name)

	var obj client.Object = &service
	r.AddResource(&obj)

	return nil
}

func getAdditionalPorts(additionalPorts []podtypes.InternalPort) []corev1.ServicePort {
	var ports []corev1.ServicePort

	for _, p := range additionalPorts {
		ports = append(ports, corev1.ServicePort{
			Name:       p.Name,
			Port:       p.Port,
			Protocol:   p.Protocol,
			TargetPort: intstr.FromInt32(p.Port),
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
