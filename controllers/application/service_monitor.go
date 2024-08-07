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
	"strings"
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

	if !r.IsIstioEnabledForNamespace(ctx, application.Namespace) {
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
			Endpoints: []pov1.Endpoint{
				{
					Path:       util.IstioMetricsPath,
					TargetPort: &util.IstioMetricsPortName,
					MetricRelabelConfigs: []pov1.RelabelConfig{
						{
							Action:       "drop",
							Regex:        strings.Join(util.DefaultMetricDropList, "|"),
							SourceLabels: []pov1.LabelName{"__name__"},
						},
					},
				},
			},
		}

		// Remove MetricRelabelConfigs if AllowAllMetrics is set to true
		if application.Spec.Prometheus != nil && application.Spec.Prometheus.AllowAllMetrics {
			serviceMonitor.Spec.Endpoints[0].MetricRelabelConfigs = nil
		}

		return nil
	})

	r.SetControllerFinishedOutcome(ctx, application, controllerName, err)

	return util.RequeueWithError(err)
}
