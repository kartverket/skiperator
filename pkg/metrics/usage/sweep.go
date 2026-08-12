package usage

import (
	"context"

	"github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/api/v1beta1"
	"github.com/kartverket/skiperator/pkg/log"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Every computed gauge counts the same two things: namespaces and Skiperator
// CRs. Collecting that once per tick keeps the API traffic at one List per
// namespace plus one List per kind, no matter how many gauges are registered.

// sweptKinds are the Skiperator CRs the gauges count. Listed cluster-wide, not
// per namespace.
var sweptKinds = []struct {
	groupVersion schema.GroupVersion
	kind         string
}{
	{v1alpha1.GroupVersion, typeApplication},
	{v1alpha1.GroupVersion, typeRouting},
	{v1beta1.GroupVersion, typeSKIPJob},
}

// namespaceLabels holds the organizational labels copied from a namespace.
type namespaceLabels struct {
	team     string
	division string
}

// usageResource is one Skiperator CR plus the labels of the namespace it lives
// in, so gauges never have to look the namespace up again.
type usageResource struct {
	item     unstructured.Unstructured
	kind     string
	team     string
	division string
}

// clusterUsage is one snapshot of what the gauges count.
type clusterUsage struct {
	namespaces []namespaceLabels
	resources  []usageResource
}

// collectClusterUsage lists namespaces and every swept kind once. It reports
// ok=false when namespaces cannot be listed, since every gauge is keyed by
// namespace labels and stale gauge values beat wrong ones.
func collectClusterUsage(ctx context.Context, k client.Client, logger log.Logger) (clusterUsage, bool) {
	namespaces := &corev1.NamespaceList{}
	if err := k.List(ctx, namespaces); err != nil {
		logger.Error(err, "failed to list namespaces")
		return clusterUsage{}, false
	}

	usage := clusterUsage{namespaces: make([]namespaceLabels, 0, len(namespaces.Items))}
	labelsByNamespace := make(map[string]namespaceLabels, len(namespaces.Items))
	for _, ns := range namespaces.Items {
		labels := namespaceLabels{
			team:     valueOrDefault(ns.Labels[labelTeam]),
			division: valueOrDefault(ns.Labels[labelDivision]),
		}
		labelsByNamespace[ns.Name] = labels
		usage.namespaces = append(usage.namespaces, labels)
	}

	for _, swept := range sweptKinds {
		list := &unstructured.UnstructuredList{}
		list.SetGroupVersionKind(swept.groupVersion.WithKind(swept.kind + "List"))
		if err := k.List(ctx, list); err != nil {
			logger.Error(err, "failed to list resources", "kind", swept.kind)
			continue
		}
		for _, item := range list.Items {
			labels, ok := labelsByNamespace[item.GetNamespace()]
			if !ok {
				// Namespace appeared between the namespace and resource list calls;
				// fall back to defaults instead of empty label values.
				labels = namespaceLabels{team: unknownValue, division: unknownValue}
			}
			usage.resources = append(usage.resources, usageResource{
				item:     item,
				kind:     swept.kind,
				team:     labels.team,
				division: labels.division,
			})
		}
	}
	return usage, true
}

// routables iterates the CRs that carry a routing provider. SKIPJob has no
// ingress, so it is not one of them.
func (u clusterUsage) routables(fn func(resource usageResource)) {
	for _, resource := range u.resources {
		if resource.kind == typeSKIPJob {
			continue
		}
		fn(resource)
	}
}
