package gwapi

import (
	"context"
	"fmt"
	"strings"

	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/api/meta"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gatewayapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

func validateApplicationConflicts(ctx context.Context, c client.Client, application *skiperatorv1alpha1.Application) error {
	if !application.UsesStandardRouting() {
		return nil
	}
	hosts, err := application.Hostnames()
	if err != nil {
		return err
	}

	listenerSets := &gatewayapiv1.ListenerSetList{}
	if err := c.List(ctx, listenerSets); err != nil {
		return fmt.Errorf("failed to list Gateway API ListenerSets: %w", err)
	}
	for _, host := range hosts.AllHosts() {
		for _, listenerSet := range listenerSets.Items {
			if !skiperatorManaged(listenerSet.Labels) || sameApplication(listenerSet.Labels, application) {
				continue
			}
			if listenerSet.Spec.ParentRef.Name != GatewayNameForHost(host.Hostname) {
				continue
			}
			for _, listener := range listenerSet.Spec.Listeners {
				if listenerCoversHostname(listener.Hostname, host.Hostname) {
					if listenerSetAccepted(listenerSet) {
						return fmt.Errorf("hostname %q already has an accepted ListenerSet %s/%s", host.Hostname, listenerSet.Namespace, listenerSet.Name)
					}
					return fmt.Errorf("hostname %q already has a pending ListenerSet %s/%s", host.Hostname, listenerSet.Namespace, listenerSet.Name)
				}
			}
		}
	}
	return nil
}

func listenerSetAccepted(listenerSet gatewayapiv1.ListenerSet) bool {
	return meta.IsStatusConditionTrue(listenerSet.Status.Conditions, string(gatewayapiv1.ListenerSetConditionAccepted))
}

func skiperatorManaged(labels map[string]string) bool {
	return labels["app.kubernetes.io/managed-by"] == "skiperator"
}

func sameApplication(labels map[string]string, application *skiperatorv1alpha1.Application) bool {
	return labels["skiperator.kartverket.no/controller"] == "application" &&
		labels["application.skiperator.no/app-name"] == application.Name &&
		labels["application.skiperator.no/app-namespace"] == application.Namespace
}

func listenerCoversHostname(listenerHostname *gatewayapiv1.Hostname, hostname string) bool {
	return listenerHostname == nil || strings.EqualFold(string(*listenerHostname), hostname)
}
