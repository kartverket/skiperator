package servicemonitor

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
	if r.GetType() != reconciliation.ApplicationType {
		return fmt.Errorf("unsupported type %s in service monitor", r.GetType())
	}
	application, ok := r.GetSKIPObject().(*skiperatorv1alpha1.Application)
	if !ok {
		err := fmt.Errorf("failed to cast resource to application")
		ctxLog.Error(err, "Failed to generate service monitor")
		return err
	}
	ctxLog.Debug("Attempting to generate service monitor for application", "application", application.Name)

	serviceMonitor := pov1.ServiceMonitor{ObjectMeta: metav1.ObjectMeta{
		Namespace: application.Namespace,
		Name:      application.Name,
		Labels:    map[string]string{"instance": "primary"},
	}}

	if !r.IsIstioEnabled() {
		return nil
	}

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

	ctxLog.Debug("Finished generating service monitor for application", "application", application.Name)

	var obj client.Object = &serviceMonitor
	r.AddResource(&obj)
	return nil
}
