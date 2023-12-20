package applicationcontroller

import (
	"context"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/util"
	pov1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *ApplicationReconciler) reconcileServiceMonitor(ctx context.Context, application *skiperatorv1alpha1.Application) (reconcile.Result, error) {
	controllerName := "ServiceMonitor"
	r.SetControllerProgressing(ctx, application, controllerName)

	if !r.isCrdPresent(ctx, "servicemonitors.monitoring.coreos.com") {
		r.SetControllerFinishedOutcome(ctx, application, controllerName, nil)
		return util.DoNotRequeue()
	}

	serviceMonitor := pov1.ServiceMonitor{ObjectMeta: metav1.ObjectMeta{
		Namespace: application.Namespace,
		Name:      application.Name,
		Labels:    map[string]string{"instance": "primary"},
	}}

	shouldReconcile, err := r.ShouldReconcile(ctx, &serviceMonitor)
	if err != nil || !shouldReconcile {
		r.SetControllerFinishedOutcome(ctx, application, controllerName, err)
		return util.RequeueWithError(err)
	}

	if application.Spec.Prometheus == nil {
		err := client.IgnoreNotFound(r.GetClient().Delete(ctx, &serviceMonitor))
		if err != nil {
			r.SetControllerError(ctx, application, controllerName, err)
			return util.RequeueWithError(err)
		}

		r.SetControllerFinishedOutcome(ctx, application, controllerName, nil)
		return util.DoNotRequeue()
	}

	_, err = ctrlutil.CreateOrPatch(ctx, r.GetClient(), &serviceMonitor, func() error {
		// Set application as owner of the service
		err := ctrlutil.SetControllerReference(application, &serviceMonitor, r.GetScheme())
		if err != nil {
			r.SetControllerError(ctx, application, controllerName, err)
			return err
		}

		r.SetLabelsFromApplication(&serviceMonitor, *application)
		util.SetCommonAnnotations(&serviceMonitor)

		serviceMonitor.Spec = pov1.ServiceMonitorSpec{
			Selector: metav1.LabelSelector{
				MatchLabels: util.GetPodAppSelector(application.Name),
			},
			NamespaceSelector: pov1.NamespaceSelector{
				MatchNames: []string{application.Namespace},
			},
			Endpoints: r.determineEndpoint(ctx, application),
		}

		return nil
	})

	r.SetControllerFinishedOutcome(ctx, application, controllerName, err)

	return util.RequeueWithError(err)
}

func (r *ApplicationReconciler) determineEndpoint(ctx context.Context, application *skiperatorv1alpha1.Application) []pov1.Endpoint {
	ep := pov1.Endpoint{
		Path: util.IstioMetricsPath, TargetPort: &util.IstioMetricsPortName,
	}

	if r.IsIstioEnabledForNamespace(ctx, application.Namespace) {
		return []pov1.Endpoint{ep}
	}

	return []pov1.Endpoint{
		{
			Path:       application.Spec.Prometheus.Path,
			TargetPort: &application.Spec.Prometheus.Port,
		},
	}
}
