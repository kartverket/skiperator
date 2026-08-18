package sidecar

import (
	"fmt"
	"slices"

	"github.com/kartverket/skiperator/api/common/podtypes"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	skiperatorv1beta1 "github.com/kartverket/skiperator/api/v1beta1"
	"github.com/kartverket/skiperator/pkg/reconciliation"
	"github.com/kartverket/skiperator/pkg/util"
	networkingv1api "istio.io/api/networking/v1"
	networkingv1 "istio.io/client-go/pkg/apis/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type outboundRuleResolver interface {
	GetResolvedOutboundRules() []reconciliation.ResolvedOutboundRule
}

func init() {
	multiGenerator.Register(reconciliation.ApplicationType, generateForWorkload)
	multiGenerator.Register(reconciliation.JobType, generateForWorkload)
}

func generateForWorkload(r reconciliation.Reconciliation) error {
	object := r.GetSKIPObject()
	accessPolicy, gcp, image, err := getSidecarSettings(object)
	if err != nil {
		return err
	}

	resolvedOutboundRules, err := getResolvedOutboundRules(r)
	if err != nil {
		return err
	}

	hosts := getEgressHosts(accessPolicy, resolvedOutboundRules)
	if util.IsCloudSqlProxyEnabled(gcp) {
		hosts = append(hosts, fmt.Sprintf("./%s-%x.cloudsql", object.GetName(), util.GenerateHashFromName(image)))
	}

	workloadName := getWorkloadName(object.GetName(), r.GetType())
	sidecar := networkingv1.Sidecar{ObjectMeta: metav1.ObjectMeta{Namespace: object.GetNamespace(), Name: workloadName}}
	sidecar.Spec = networkingv1api.Sidecar{
		WorkloadSelector: &networkingv1api.WorkloadSelector{
			Labels: util.GetPodAppSelector(workloadName),
		},
		Egress: []*networkingv1api.IstioEgressListener{
			{
				Hosts: normalizeHosts(hosts),
			},
		},
		OutboundTrafficPolicy: &networkingv1api.OutboundTrafficPolicy{
			Mode: networkingv1api.OutboundTrafficPolicy_REGISTRY_ONLY,
		},
	}

	r.AddResource(&sidecar)
	return nil
}

func getResolvedOutboundRules(r reconciliation.Reconciliation) ([]reconciliation.ResolvedOutboundRule, error) {
	resolver, ok := r.(outboundRuleResolver)
	if !ok {
		return nil, fmt.Errorf("reconciliation type %s does not resolve outbound rules", r.GetType())
	}

	return resolver.GetResolvedOutboundRules(), nil
}

// Avoid GetCommonSpec here because it can dereference fields that sidecar generation does not need.
// For example, Application.GetCommonSpec reads a.Spec.IstioSettings.IstioSettingsBase, which can panic
// when IstioSettings is nil even though this generator only needs AccessPolicy, GCP, and Image.
func getSidecarSettings(object skiperatorv1alpha1.SKIPObject) (*podtypes.AccessPolicy, *podtypes.GCP, string, error) {
	switch typedObject := object.(type) {
	case *skiperatorv1alpha1.Application:
		return typedObject.Spec.AccessPolicy, typedObject.Spec.GCP, typedObject.Spec.Image, nil
	case *skiperatorv1beta1.SKIPJob:
		return typedObject.Spec.AccessPolicy, typedObject.Spec.GCP, typedObject.Spec.Image, nil
	default:
		return nil, nil, "", fmt.Errorf("unsupported workload type %T for sidecar generation", object)
	}
}

func getEgressHosts(accessPolicy *podtypes.AccessPolicy, resolvedOutboundRules []reconciliation.ResolvedOutboundRule) []string {
	if accessPolicy == nil {
		return nil
	}

	hosts := make([]string, 0)
	if accessPolicy.Outbound != nil {
		hosts = make([]string, 0, len(accessPolicy.Outbound.External)+len(resolvedOutboundRules))
		for _, rule := range accessPolicy.Outbound.External {
			if rule.Host != "" {
				hosts = append(hosts, fmt.Sprintf("./%s", rule.Host))
			}
		}
	}

	for _, rule := range resolvedOutboundRules {
		for _, namespace := range rule.Namespaces {
			hosts = append(hosts, getServiceHost(namespace, rule.Application))
		}
	}

	return hosts
}

func getServiceHost(namespace string, application string) string {
	return fmt.Sprintf("%s/%s.%s.svc.cluster.local", namespace, application, namespace)
}

func getWorkloadName(name string, resourceType reconciliation.ObjectType) string {
	if resourceType == reconciliation.JobType {
		return fmt.Sprintf("%s-skipjob", name)
	}

	return name
}

func normalizeHosts(hosts []string) []string {
	if len(hosts) == 0 {
		// Istio uses "~/*" to import no services into the workload sidecar.
		return []string{"~/*"}
	}

	slices.Sort(hosts)
	return slices.Compact(hosts)
}
