package gatewayapi

import (
	"fmt"

	"github.com/kartverket/skiperator/api/common/istiotypes"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/reconciliation"
	gatewayapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

func init() {
	multiGenerator.Register(reconciliation.ApplicationType, generateForApplication)
}

func generateForApplication(r reconciliation.Reconciliation) error {
	ctxLog := r.GetLogger()
	ctxLog.Debug("Attempting to generate gateway api resources for application", "application", r.GetSKIPObject().GetName())

	application, ok := r.GetSKIPObject().(*skiperatorv1alpha1.Application)
	if !ok {
		return fmt.Errorf("failed to cast object to Application")
	}
	if !application.UsesStandardRouting() {
		return nil
	}

	hosts, err := application.Spec.Hosts()
	if err != nil {
		return err
	}

	listenerSetNames, hostnames, err := addListenerSets(r, application.Namespace, application.Name, hosts, application.GetCertificateName)
	if err != nil {
		return err
	}

	redirectToHTTPS := application.Spec.RedirectToHTTPS != nil && *application.Spec.RedirectToHTTPS
	if redirectToHTTPS {
		r.AddResource(newRedirectRoute(application.Namespace, application.Name, "", listenerSetNames, hostnames))
	}

	backend, err := backendRule("default-app-route", int32(application.Spec.Port), "/", false, applicationRetries(application), func(field string, value string) {
		ctxLog.Warn("Ignoring unsupported Gateway API retry option", "kind", "Application", "namespace", application.Namespace, "name", application.Name, "field", field, "value", value)
	})
	if err != nil {
		return err
	}
	backend.BackendRefs[0].Name = gatewayapiv1.ObjectName(application.Name)

	r.AddResource(newBackendRoute(application.Namespace, application.Name, "", listenerSetNames, hostnames, []gatewayapiv1.HTTPRouteRule{backend}))

	ctxLog.Debug("Finished generating gateway api resources for application", "application", application.Name)
	return nil
}

func applicationRetries(application *skiperatorv1alpha1.Application) *istiotypes.Retries {
	if application.Spec.IstioSettings == nil {
		return nil
	}
	return application.Spec.IstioSettings.Retries
}
