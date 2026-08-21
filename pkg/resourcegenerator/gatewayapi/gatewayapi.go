package gatewayapi

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/kartverket/skiperator/api/common"
	"github.com/kartverket/skiperator/api/common/istiotypes"
	"github.com/kartverket/skiperator/pkg/gwapi"
	"github.com/kartverket/skiperator/pkg/reconciliation"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/resourceutils/generator"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	gatewayapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

const (
	httpSectionName  gatewayapiv1.SectionName = "http"
	httpsSectionName gatewayapiv1.SectionName = "https"
)

type unsupportedRetryOptionFunc func(field string, value string)

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

// retriable5xxCodes expands Istio's "5xx" retry shorthand. Gateway API takes
// explicit status codes, so the shorthand becomes the 5xx codes the Retries
// enum permits (509 is not one of them).
var retriable5xxCodes = []int{500, 501, 502, 503, 504, 505, 506, 507, 508, 510, 511}

// retriable4xxCode is the only 4xx code Istio's "retriable-4xx" retries on.
const retriable4xxCode = 409

// applyRetries translates Istio retry settings onto a Gateway API rule.
//
// Unreachable while the Application CRD refuses retries together with
// spec.routingProvider=Standard, because Gateway API serves rules[].retry only
// on its experimental channel. Kept ready for the day retry graduates to the
// standard channel: drop that CEL rule in api/v1alpha1/application_types.go and
// this translation takes over. Timeouts are standard channel already.
// Attempts and status codes become HTTPRouteRetry, and perTryTimeout becomes the
// per-attempt Timeouts.BackendRequest. Settings that have no Gateway API
// equivalent are reported through onUnsupportedRetryOption so users can see that
// standard routing ignored part of their legacy config.
func applyRetries(rule *gatewayapiv1.HTTPRouteRule, retries *istiotypes.Retries, onUnsupportedRetryOption unsupportedRetryOptionFunc) error {
	if retries == nil {
		return nil
	}

	attempts := 2
	if retries.Attempts != nil {
		attempts = int(*retries.Attempts)
	}
	rule.Retry = &gatewayapiv1.HTTPRouteRetry{Attempts: &attempts}

	if retries.PerTryTimeout != nil {
		timeout, err := gatewayAPIDuration(retries.PerTryTimeout.Duration)
		if err != nil {
			// Istio accepts durations GEP-2257 cannot express, e.g. sub-millisecond.
			onUnsupportedRetryOption("perTryTimeout", retries.PerTryTimeout.Duration.String())
		} else {
			rule.Timeouts = &gatewayapiv1.HTTPRouteTimeouts{BackendRequest: &timeout}
		}
	}

	if retries.RetryOnHttpResponseCodes == nil {
		return nil
	}

	codes := make([]gatewayapiv1.HTTPRouteRetryStatusCode, 0, len(*retries.RetryOnHttpResponseCodes))
	for _, code := range *retries.RetryOnHttpResponseCodes {
		values, err := retryCodes(code, onUnsupportedRetryOption)
		if err != nil {
			return err
		}
		for _, value := range values {
			codes = append(codes, gatewayapiv1.HTTPRouteRetryStatusCode(value))
		}
	}
	// retry.codes is listType=set from Gateway API 1.6, so duplicates are
	// rejected by the API server. Expanding "5xx" next to an explicit 503 is a
	// realistic way to produce them.
	slices.Sort(codes)
	codes = slices.Compact(codes)
	if len(codes) > 0 {
		rule.Retry.Codes = codes
	}
	return nil
}

// retryCodes resolves one configured retry code to explicit status codes. The
// "5xx" and "retriable-4xx" shorthands expand; anything else non-numeric is
// reported as unsupported and dropped.
func retryCodes(code intstr.IntOrString, onUnsupportedRetryOption unsupportedRetryOptionFunc) ([]int, error) {
	if code.Type == intstr.Int {
		value, err := validateRetryCode(code.IntValue())
		if err != nil {
			return nil, err
		}
		return []int{value}, nil
	}
	switch code.StrVal {
	case "5xx":
		return retriable5xxCodes, nil
	case "retriable-4xx":
		return []int{retriable4xxCode}, nil
	}
	value, err := strconv.Atoi(code.StrVal)
	if err != nil {
		onUnsupportedRetryOption("retryOnHttpResponseCodes", code.StrVal)
		return nil, nil
	}
	value, err = validateRetryCode(value)
	if err != nil {
		return nil, err
	}
	return []int{value}, nil
}

func validateRetryCode(code int) (int, error) {
	if code < 400 || code > 599 {
		return 0, fmt.Errorf("gateway api retry status code must be between 400 and 599, got %d", code)
	}
	return code, nil
}

// gatewayAPIDuration renders d in the GEP-2257 duration subset Gateway API
// accepts: whole milliseconds, largest unit first, no zero components, at most
// five digits per component. Istio accepts durations outside that subset, so
// callers must handle the error rather than assume every timeout translates.
// See https://gateway-api.sigs.k8s.io/geps/gep-2257/.
func gatewayAPIDuration(d time.Duration) (gatewayapiv1.Duration, error) {
	if d <= 0 {
		return "", fmt.Errorf("duration must be positive, got %s", d)
	}
	if d%time.Millisecond != 0 {
		return "", fmt.Errorf("cannot express sub-millisecond precision of %s", d)
	}

	units := []struct {
		suffix       string
		milliseconds int64
	}{{"h", 3600000}, {"m", 60000}, {"s", 1000}, {"ms", 1}}

	remaining := d.Milliseconds()
	var out strings.Builder
	for _, unit := range units {
		count := remaining / unit.milliseconds
		if count == 0 {
			continue
		}
		if count > 99999 {
			return "", fmt.Errorf("duration %s is larger than GEP-2257 can express", d)
		}
		fmt.Fprintf(&out, "%d%s", count, unit.suffix)
		remaining %= unit.milliseconds
	}
	return gatewayapiv1.Duration(out.String()), nil
}
