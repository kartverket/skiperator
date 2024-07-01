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
	"strings"
)

func (r *SKIPJobReconciler) reconcilePodMonitor(ctx context.Context, skipJob *skiperatorv1alpha1.SKIPJob) (reconcile.Result, error) {
	podMonitor := pov1.PodMonitor{ObjectMeta: metav1.ObjectMeta{
		Name:      skipJob.Name + "-monitor",
		Namespace: skipJob.Namespace,
		Labels:    map[string]string{"instance": "primary"},
	}}

	shouldReconcile, err := r.ShouldReconcile(ctx, &podMonitor)
	if err != nil || !shouldReconcile {
		return util.RequeueWithError(err)
	}

	if skipJob.Spec.Prometheus == nil {
		err := client.IgnoreNotFound(r.GetClient().Delete(ctx, &podMonitor))
		if err != nil {
			return util.RequeueWithError(err)
		}
		return util.DoNotRequeue()
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
			PodMetricsEndpoints: []pov1.PodMetricsEndpoint{
				{
					Path:       util.IstioMetricsPath,
					TargetPort: &util.IstioMetricsPortName,
				},
			},
		}
		if !skipJob.Spec.Prometheus.AllowAllMetrics {
			podMonitor.Spec.PodMetricsEndpoints[0].MetricRelabelConfigs = []pov1.RelabelConfig{
				{
					Action:       "drop",
					Regex:        strings.Join(util.DefaultMetricDropList, "|"),
					SourceLabels: []pov1.LabelName{"__name__"},
				},
			}
		}
		return nil
	})
	return util.RequeueWithError(err)
}
