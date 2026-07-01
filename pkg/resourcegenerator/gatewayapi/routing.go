package gatewayapi

import (
	"fmt"

	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/gwapi"
	"github.com/kartverket/skiperator/pkg/reconciliation"
	gatewayapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

func init() {
	multiGenerator.Register(reconciliation.RoutingType, generateForRouting)
}

func generateForRouting(r reconciliation.Reconciliation) error {
	ctxLog := r.GetLogger()
	ctxLog.Debug("Attempting to generate gateway api resources for routing", "routing", r.GetSKIPObject().GetName())

	routing, ok := r.GetSKIPObject().(*skiperatorv1alpha1.Routing)
	if !ok {
		return fmt.Errorf("failed to cast object to Routing")
	}
	if !routing.UsesStandardRouting() {
		return nil
	}

	hosts, err := routing.Hostnames()
	if err != nil {
		return err
	}
	// Qualify the name so a standalone Routing and an equally named Application
	// in the same namespace do not collide on Gateway API resource names.
	routePrefix := gwapi.RoutingResourcePrefix(routing.Name)
	var listenerSetNames []string
	var hostnames []gatewayapiv1.Hostname
	listenerSetNamespace := ""
	if routing.UsesSharedOwnership() {
		listenerSetNamespace = gwapi.IstioGatewayNamespace
		listenerSetNames, hostnames, err = addSharedListenerSets(r, hosts, routing.GetCertificateName)
	} else {
		listenerSetNames, hostnames, err = addListenerSets(r, routing.Namespace, routePrefix, hosts, routing.GetCertificateName)
	}
	if err != nil {
		return err
	}

	if routing.GetRedirectToHTTPS() {
		if routing.UsesSharedOwnership() {
			host := hosts.AllHosts()[0].Hostname
			r.AddResource(newRedirectRouteWithName(gwapi.IstioGatewayNamespace, gwapi.SharedRedirectRouteName(host), "", listenerSetNames, hostnames))
		} else {
			r.AddResource(newRedirectRoute(routing.Namespace, routePrefix, listenerSetNamespace, listenerSetNames, hostnames))
		}
	}

	rules := make([]gatewayapiv1.HTTPRouteRule, 0, len(routing.Spec.Routes))
	for _, route := range routing.Spec.Routes {
		rule, err := backendRule(route.TargetApp, route.Port, route.PathPrefix, route.RewriteUri, nil, func(field string, value string) {
			ctxLog.Warn("Ignoring unsupported Gateway API retry option", "kind", "Routing", "namespace", routing.Namespace, "name", routing.Name, "field", field, "value", value)
		})
		if err != nil {
			return err
		}
		rules = append(rules, rule)
	}

	r.AddResource(newBackendRoute(routing.Namespace, routePrefix, listenerSetNamespace, listenerSetNames, hostnames, rules))

	ctxLog.Debug("Finished generating gateway api resources for routing", "routing", routing.Name)
	return nil
}
