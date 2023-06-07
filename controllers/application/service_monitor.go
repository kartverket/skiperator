package applicationcontroller

import (
	"context"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/util"
	pov1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const MetricsPortName = "metrics"

func (r *ApplicationReconciler) reconcileServiceMonitor(ctx context.Context, application *skiperatorv1alpha1.Application) (reconcile.Result, error) {
	controllerName := "ServiceMonitor"
	r.SetControllerProgressing(ctx, application, controllerName)

	if !r.isCrdPresent(ctx, "servicemonitors.monitoring.coreos.com") {
		r.SetControllerFinishedOutcome(ctx, application, controllerName, nil)
		return reconcile.Result{}, nil
	}

	if !hasMetricsPort(application) {
		r.SetControllerFinishedOutcome(ctx, application, controllerName, nil)
		return reconcile.Result{}, nil
	}

	serviceMonitor := pov1.ServiceMonitor{ObjectMeta: metav1.ObjectMeta{Namespace: application.Namespace, Name: application.Name}}
	_, err := ctrlutil.CreateOrPatch(ctx, r.GetClient(), &serviceMonitor, func() error {
		// Set application as owner of the service
		err := ctrlutil.SetControllerReference(application, &serviceMonitor, r.GetScheme())
		if err != nil {
			r.SetControllerError(ctx, application, controllerName, err)
			return err
		}

		r.SetLabelsFromApplication(ctx, &serviceMonitor, *application)
		util.SetCommonAnnotations(&serviceMonitor)

		serviceMonitor.Spec = pov1.ServiceMonitorSpec{
			Selector: metav1.LabelSelector{
				MatchLabels: util.GetApplicationSelector(application.Name),
			},
			NamespaceSelector: pov1.NamespaceSelector{
				MatchNames: []string{application.Namespace},
			},
			Endpoints: []pov1.Endpoint{
				{Port: "metrics"},
			},
		}

		return nil
	})

	r.SetControllerFinishedOutcome(ctx, application, controllerName, err)

	return reconcile.Result{}, err
}

func hasMetricsPort(application *skiperatorv1alpha1.Application) bool {
	for _, p := range application.Spec.AdditionalPorts {
		if p.Name == MetricsPortName {
			return true
		}
	}

	return false
}
