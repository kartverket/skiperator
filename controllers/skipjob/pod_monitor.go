package skipjobcontroller

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

func (r *SKIPJobReconciler) reconcilePodMonitor(ctx context.Context, skipJob *skiperatorv1alpha1.SKIPJob) (reconcile.Result, error) {
	podMonitor := pov1.PodMonitor{ObjectMeta: metav1.ObjectMeta{
		Name:      skipJob.Name,
		Namespace: skipJob.Namespace,
		Labels:    map[string]string{"instance": "primary"},
	}}

	shouldReconcile, err := r.ShouldReconcile(ctx, &podMonitor)
	if err != nil || !shouldReconcile {
		return reconcile.Result{}, err
	}

	if skipJob.Spec.Prometheus == nil {
		err := client.IgnoreNotFound(r.GetClient().Delete(ctx, &podMonitor))
		if err != nil {
			return reconcile.Result{}, err
		}
		return reconcile.Result{}, nil
	}

	_, err = ctrlutil.CreateOrPatch(ctx, r.GetClient(), &podMonitor, func() error {
		err := ctrlutil.SetControllerReference(skipJob, &podMonitor, r.GetScheme())
		if err != nil {
			return err
		}
		podMonitor.Spec = pov1.PodMonitorSpec{
			Selector: metav1.LabelSelector{
				MatchLabels: util.GetPodAppSelector(skipJob.Name),
			},
			NamespaceSelector: pov1.NamespaceSelector{
				MatchNames: []string{skipJob.Namespace},
			},
			PodMetricsEndpoints: r.determineEndpoint(ctx, skipJob),
		}
		return nil
	})
	return reconcile.Result{}, err
}

func (r *SKIPJobReconciler) determineEndpoint(ctx context.Context, application *skiperatorv1alpha1.SKIPJob) []pov1.PodMetricsEndpoint {
	ep := pov1.PodMetricsEndpoint{
		Path: util.IstioMetricsPath, TargetPort: &util.IstioMetricsPortName,
	}
	if r.IsIstioEnabledForNamespace(ctx, application.Namespace) {
		return []pov1.PodMetricsEndpoint{ep}
	}
	return []pov1.PodMetricsEndpoint{
		{
			Path:       application.Spec.Prometheus.Path,
			TargetPort: &application.Spec.Prometheus.Port,
		},
	}
}
