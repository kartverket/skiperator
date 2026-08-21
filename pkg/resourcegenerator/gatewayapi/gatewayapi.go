package gatewayapi

import (
	"github.com/kartverket/skiperator/api/common"
	"github.com/kartverket/skiperator/pkg/gwapi"
	"github.com/kartverket/skiperator/pkg/reconciliation"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/resourceutils/generator"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gatewayapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

const (
	httpSectionName  gatewayapiv1.SectionName = "http"
	httpsSectionName gatewayapiv1.SectionName = "https"
)

var multiGenerator = generator.NewMulti()

// Generate creates Kubernetes Gateway API resources for Applications and
// Routings that opt into the standard routing provider.
func Generate(r reconciliation.Reconciliation) error {
	return multiGenerator.Generate(r, "Gateway API")
}

// parentGatewayRef points a ListenerSet at the shared Gateway selected for the
// hostname. ListenerSets live in application namespaces, while shared Gateways
// live in istio-gateways.
func parentGatewayRef(hostname string) gatewayapiv1.ParentGatewayReference {
	namespace := gatewayapiv1.Namespace(gwapi.IstioGatewayNamespace)
	return gatewayapiv1.ParentGatewayReference{
		Name:      gwapi.GatewayNameForHost(hostname),
		Namespace: &namespace,
	}
}

// parentListenerSetRef points an HTTPRoute at one ListenerSet listener. The
// section selects whether the route is attached to the HTTP or HTTPS listener.
func parentListenerSetRef(namespace string, name string, section gatewayapiv1.SectionName) gatewayapiv1.ParentReference {
	group := gatewayapiv1.Group(gatewayapiv1.GroupName)
	kind := gatewayapiv1.Kind("ListenerSet")
	ref := gatewayapiv1.ParentReference{
		Group:       &group,
		Kind:        &kind,
		Name:        gatewayapiv1.ObjectName(name),
		SectionName: &section,
	}
	if namespace != "" {
		ref.Namespace = new(gatewayapiv1.Namespace(namespace))
	}
	return ref
}

// newListenerSet adds HTTP and HTTPS listeners for one hostname. TLS
// termination happens on the HTTPS listener using a Secret in the same namespace
// as the ListenerSet.
func newListenerSet(namespace string, name string, hostname string, secretName string) *gatewayapiv1.ListenerSet {
	return &gatewayapiv1.ListenerSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: gatewayapiv1.ListenerSetSpec{
			ParentRef: parentGatewayRef(hostname),
			Listeners: listeners(hostname, secretName),
		},
	}
}

// parentRefs expands ListenerSet names into ParentRefs for one listener
// section. Applications can have several ListenerSets because they can expose
// several hostnames.
func parentRefs(listenerSetNamespace string, listenerSetNames []string, section gatewayapiv1.SectionName) []gatewayapiv1.ParentReference {
	parents := make([]gatewayapiv1.ParentReference, 0, len(listenerSetNames))
	for _, name := range listenerSetNames {
		parents = append(parents, parentListenerSetRef(listenerSetNamespace, name, section))
	}
	return parents
}

func secretRef(name string) gatewayapiv1.SecretObjectReference {
	return gatewayapiv1.SecretObjectReference{
		Name: gatewayapiv1.ObjectName(name),
	}
}

// addListenerSets creates one ListenerSet per hostname and returns the names
// and hostnames needed when building HTTPRoutes for those listeners.
func addListenerSets(r reconciliation.Reconciliation, namespace string, prefix string, hosts common.HostCollection, certificateName func(*common.Host) (string, error)) ([]string, []gatewayapiv1.Hostname, error) {
	listenerSetNames := make([]string, 0, hosts.Count())
	hostnames := make([]gatewayapiv1.Hostname, 0, hosts.Count())

	for _, h := range hosts.AllHosts() {
		name := gwapi.ListenerSetName(prefix, h.Hostname)
		secretName, err := certificateName(h)
		if err != nil {
			return nil, nil, err
		}
		listenerSetNames = append(listenerSetNames, name)
		hostnames = append(hostnames, gatewayapiv1.Hostname(h.Hostname))
		r.AddResource(newListenerSet(namespace, name, h.Hostname, secretName))
	}
	return listenerSetNames, hostnames, nil
}

// newRedirectRoute creates the HTTP listener route that sends clients to HTTPS.
func newRedirectRoute(namespace string, prefix string, listenerSetNamespace string, listenerSetNames []string, hostnames []gatewayapiv1.Hostname) *gatewayapiv1.HTTPRoute {
	return newHTTPRoute(namespace, gwapi.RedirectRouteName(prefix), listenerSetNamespace, listenerSetNames, httpSectionName, hostnames, []gatewayapiv1.HTTPRouteRule{redirectRule()})
}

// newBackendRoute creates the HTTPS listener route that sends traffic to
// Kubernetes Services.
func newBackendRoute(namespace string, prefix string, listenerSetNamespace string, listenerSetNames []string, hostnames []gatewayapiv1.Hostname, rules []gatewayapiv1.HTTPRouteRule) *gatewayapiv1.HTTPRoute {
	// Backend HTTPRoute uses the bare object name.
	return newHTTPRoute(namespace, prefix, listenerSetNamespace, listenerSetNames, httpsSectionName, hostnames, rules)
}

// newHTTPRoute builds a route attached to a set of ListenerSets on one listener
// section. Rules decide whether the route redirects or forwards to backends.
func newHTTPRoute(namespace string, name string, listenerSetNamespace string, listenerSetNames []string, section gatewayapiv1.SectionName, hostnames []gatewayapiv1.Hostname, rules []gatewayapiv1.HTTPRouteRule) *gatewayapiv1.HTTPRoute {
	return &gatewayapiv1.HTTPRoute{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: gatewayapiv1.HTTPRouteSpec{
			CommonRouteSpec: gatewayapiv1.CommonRouteSpec{
				ParentRefs: parentRefs(listenerSetNamespace, listenerSetNames, section),
			},
			Hostnames: hostnames,
			Rules:     rules,
		},
	}
}

// listeners returns the two listeners Skiperator exposes for each hostname:
// port 80 HTTP for redirects and port 443 HTTPS for backend routes.
func listeners(hostname string, secretName string) []gatewayapiv1.ListenerEntry {
	terminate := gatewayapiv1.TLSModeTerminate

	return []gatewayapiv1.ListenerEntry{
		{
			Name:     httpSectionName,
			Hostname: new(gatewayapiv1.Hostname(hostname)),
			Port:     gatewayapiv1.PortNumber(80),
			Protocol: gatewayapiv1.HTTPProtocolType,
		},
		{
			Name:     httpsSectionName,
			Hostname: new(gatewayapiv1.Hostname(hostname)),
			Port:     gatewayapiv1.PortNumber(443),
			Protocol: gatewayapiv1.HTTPSProtocolType,
			TLS: &gatewayapiv1.ListenerTLSConfig{
				Mode:            &terminate,
				CertificateRefs: []gatewayapiv1.SecretObjectReference{secretRef(secretName)},
			},
		},
	}
}

// redirectRule returns a Gateway API equivalent of the legacy Istio
// redirect-to-https rule.
func redirectRule() gatewayapiv1.HTTPRouteRule {
	scheme := "https"
	statusCode := 308
	return gatewayapiv1.HTTPRouteRule{
		Filters: []gatewayapiv1.HTTPRouteFilter{
			{
				Type: gatewayapiv1.HTTPRouteFilterRequestRedirect,
				RequestRedirect: &gatewayapiv1.HTTPRequestRedirectFilter{
					Scheme:     &scheme,
					StatusCode: &statusCode,
				},
			},
		},
	}
}

// backendRule returns one Gateway API HTTPRoute rule for a backend Service.
// Routing objects may create several such rules, while Application creates one
// default rule pointing to the Application Service. Retries are set by the
// caller, since only Application supports them.
func backendRule(name string, serviceName string, port int32, pathPrefix string, rewrite bool) gatewayapiv1.HTTPRouteRule {
	portNumber := gatewayapiv1.PortNumber(port)
	pathType := gatewayapiv1.PathMatchPathPrefix
	ruleName := gatewayapiv1.SectionName(name)
	rule := gatewayapiv1.HTTPRouteRule{
		Name: &ruleName,
		Matches: []gatewayapiv1.HTTPRouteMatch{
			{
				Path: &gatewayapiv1.HTTPPathMatch{
					Type:  &pathType,
					Value: &pathPrefix,
				},
			},
		},
		BackendRefs: []gatewayapiv1.HTTPBackendRef{
			{
				BackendRef: gatewayapiv1.BackendRef{
					BackendObjectReference: gatewayapiv1.BackendObjectReference{
						Name: gatewayapiv1.ObjectName(serviceName),
						Port: &portNumber,
					},
				},
			},
		},
	}
	if rewrite {
		replace := "/"
		rule.Filters = []gatewayapiv1.HTTPRouteFilter{
			{
				Type: gatewayapiv1.HTTPRouteFilterURLRewrite,
				URLRewrite: &gatewayapiv1.HTTPURLRewriteFilter{
					Path: &gatewayapiv1.HTTPPathModifier{
						Type:               gatewayapiv1.PrefixMatchHTTPPathModifier,
						ReplacePrefixMatch: &replace,
					},
				},
			},
		}
	}
	return rule
}
