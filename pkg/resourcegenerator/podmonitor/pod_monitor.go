package podmonitor

import (
	"fmt"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/reconciliation"
	"github.com/kartverket/skiperator/pkg/util"
	pov1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

func Generate(r reconciliation.Reconciliation) error {
	ctxLog := r.GetLogger()
	ctxLog.Debug("Attempting to generate podmonitor for skipjob", "skipjob", r.GetSKIPObject().GetName())

	if r.GetType() != reconciliation.JobType {
		return fmt.Errorf("podmonitor only supports skipjob type, got %s", r.GetType())
	}

	skipJob := r.GetSKIPObject().(*skiperatorv1alpha1.SKIPJob)

	if skipJob.Spec.Prometheus == nil {
		return nil
	}

	podMonitor := pov1.PodMonitor{ObjectMeta: metav1.ObjectMeta{
		Name:      skipJob.Name + "-monitor",
		Namespace: skipJob.Namespace,
		Labels:    map[string]string{"instance": "primary"},
	}}

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
	var obj client.Object = &podMonitor
	r.AddResource(&obj)

	ctxLog.Debug("Finished generating configmap", "name", skipJob.GetName())
	return nil
}
