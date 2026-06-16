package gwapi

import (
	"context"
	"testing"

	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	certmanagermetav1 "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
	commontypes "github.com/kartverket/skiperator/api/common"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/resourceschemas"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	istionetworkingv1 "istio.io/client-go/pkg/apis/networking/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	gatewayapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

func TestApplicationHostnameConflict(t *testing.T) {
	scheme := runtime.NewScheme()
	resourceschemas.AddSchemas(scheme)
	listenerSet := &gatewayapiv1.ListenerSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "accepted",
			Namespace: "team-b",
			Labels: map[string]string{
				"app.kubernetes.io/managed-by":        "skiperator",
				"skiperator.kartverket.no/controller": "application",
			},
		},
		Spec: gatewayapiv1.ListenerSetSpec{
			ParentRef: gatewayapiv1.ParentGatewayReference{Name: "istio-external"},
			Listeners: []gatewayapiv1.ListenerEntry{
				{Hostname: gatewayHostname("app.example.com")},
			},
		},
		Status: gatewayapiv1.ListenerSetStatus{
			Conditions: []metav1.Condition{{Type: string(gatewayapiv1.ListenerSetConditionAccepted), Status: metav1.ConditionTrue}},
		},
	}
	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(listenerSet).Build()
	application := &skiperatorv1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{Name: "app", Namespace: "team-a"},
		Spec: skiperatorv1alpha1.ApplicationSpec{
			Ingresses:       []string{"app.example.com"},
			RoutingProvider: skiperatorv1alpha1.RoutingProviderStandard,
		},
	}

	err := ValidateConflicts(context.Background(), c, application)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "already has an accepted ListenerSet")
}

func TestApplicationHostnameConflictWithWildcardListener(t *testing.T) {
	scheme := runtime.NewScheme()
	resourceschemas.AddSchemas(scheme)
	listenerSet := &gatewayapiv1.ListenerSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "accepted",
			Namespace: "team-b",
			Labels: map[string]string{
				"app.kubernetes.io/managed-by":        "skiperator",
				"skiperator.kartverket.no/controller": "application",
			},
		},
		Spec: gatewayapiv1.ListenerSetSpec{
			ParentRef: gatewayapiv1.ParentGatewayReference{Name: "istio-external"},
			Listeners: []gatewayapiv1.ListenerEntry{{}},
		},
		Status: gatewayapiv1.ListenerSetStatus{
			Conditions: []metav1.Condition{{Type: string(gatewayapiv1.ListenerSetConditionAccepted), Status: metav1.ConditionTrue}},
		},
	}
	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(listenerSet).Build()
	application := &skiperatorv1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{Name: "app", Namespace: "team-a"},
		Spec: skiperatorv1alpha1.ApplicationSpec{
			Ingresses:       []string{"app.example.com"},
			RoutingProvider: skiperatorv1alpha1.RoutingProviderStandard,
		},
	}

	err := ValidateConflicts(context.Background(), c, application)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "already has an accepted ListenerSet")
}

func TestApplicationHostnameConflictWithPendingListenerSet(t *testing.T) {
	scheme := runtime.NewScheme()
	resourceschemas.AddSchemas(scheme)
	listenerSet := &gatewayapiv1.ListenerSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pending",
			Namespace: "team-b",
			Labels: map[string]string{
				"app.kubernetes.io/managed-by":        "skiperator",
				"skiperator.kartverket.no/controller": "application",
			},
		},
		Spec: gatewayapiv1.ListenerSetSpec{
			ParentRef: gatewayapiv1.ParentGatewayReference{Name: "istio-external"},
			Listeners: []gatewayapiv1.ListenerEntry{
				{Hostname: gatewayHostname("app.example.com")},
			},
		},
	}
	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(listenerSet).Build()
	application := &skiperatorv1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{Name: "app", Namespace: "team-a"},
		Spec: skiperatorv1alpha1.ApplicationSpec{
			Ingresses:       []string{"app.example.com"},
			RoutingProvider: skiperatorv1alpha1.RoutingProviderStandard,
		},
	}

	err := ValidateConflicts(context.Background(), c, application)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "pending ListenerSet")
}

func TestApplicationStandardRoutingKeepsLegacyUntilReady(t *testing.T) {
	scheme := runtime.NewScheme()
	resourceschemas.AddSchemas(scheme)
	application := gatewayAPIApplication()
	legacy := &istionetworkingv1.VirtualService{ObjectMeta: metav1.ObjectMeta{Name: "app-ingress", Namespace: "team-a"}}
	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(application, legacy).Build()

	state, err := EvaluateRoutingState(context.Background(), c, application, application.GetStatus())
	require.NoError(t, err)

	assert.True(t, state.GenerateLegacyRouting)
	assert.False(t, state.Readiness.Ready)
	assert.Contains(t, state.Readiness.Message, "Certificate")
}

func TestApplicationStandardRoutingPrunesLegacyWhenReady(t *testing.T) {
	scheme := runtime.NewScheme()
	resourceschemas.AddSchemas(scheme)
	application := gatewayAPIApplication()
	certificateName, err := application.GetCertificateName(mustHost(t, "app.example.com"))
	require.NoError(t, err)
	objects := []client.Object{
		application,
		&istionetworkingv1.VirtualService{ObjectMeta: metav1.ObjectMeta{Name: "app-ingress", Namespace: "team-a"}},
		readyGateway(IstioGatewayNamespace, ExternalGatewayName),
		readyCertificate("team-a", certificateName),
		tlsSecret("team-a", certificateName),
		readyListenerSet("team-a", ListenerSetName("app", "app.example.com")),
		readyHTTPRoute("team-a", "app"),
	}
	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(objects...).Build()

	state, err := EvaluateRoutingState(context.Background(), c, application, application.GetStatus())
	require.NoError(t, err)

	assert.False(t, state.GenerateLegacyRouting)
	assert.True(t, state.Readiness.Ready)
}

func TestApplicationStandardRoutingGreenfieldSkipsLegacy(t *testing.T) {
	scheme := runtime.NewScheme()
	resourceschemas.AddSchemas(scheme)
	application := gatewayAPIApplication()
	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(application).Build()

	state, err := EvaluateRoutingState(context.Background(), c, application, application.GetStatus())
	require.NoError(t, err)

	assert.False(t, state.GenerateLegacyRouting)
	assert.False(t, state.Readiness.Ready)
}

func TestApplicationStandardRoutingCustomCertRequiresSecret(t *testing.T) {
	scheme := runtime.NewScheme()
	resourceschemas.AddSchemas(scheme)
	application := gatewayAPIApplication()
	application.Spec.Ingresses = []string{"app.example.com+custom-tls"}
	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(application).Build()

	state, err := EvaluateRoutingState(context.Background(), c, application, application.GetStatus())
	require.NoError(t, err)

	assert.False(t, state.Readiness.Ready)
	assert.Contains(t, state.Readiness.Message, "Secret team-a/custom-tls")
}

func TestRoutingPathConflict(t *testing.T) {
	scheme := runtime.NewScheme()
	resourceschemas.AddSchemas(scheme)
	pathType := gatewayapiv1.PathMatchPathPrefix
	path := "/v1"
	route := &gatewayapiv1.HTTPRoute{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "accepted",
			Namespace: "team-b",
			Labels: map[string]string{
				"app.kubernetes.io/managed-by":        "skiperator",
				"skiperator.kartverket.no/controller": "routing",
			},
		},
		Spec: gatewayapiv1.HTTPRouteSpec{
			Hostnames: []gatewayapiv1.Hostname{"API.example.COM"},
			Rules: []gatewayapiv1.HTTPRouteRule{
				{Matches: []gatewayapiv1.HTTPRouteMatch{{Path: &gatewayapiv1.HTTPPathMatch{Type: &pathType, Value: &path}}}},
			},
		},
		Status: gatewayapiv1.HTTPRouteStatus{
			RouteStatus: gatewayapiv1.RouteStatus{
				Parents: []gatewayapiv1.RouteParentStatus{
					{Conditions: []metav1.Condition{{Type: string(gatewayapiv1.RouteConditionAccepted), Status: metav1.ConditionTrue}}},
				},
			},
		},
	}
	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(route).Build()
	routing := &skiperatorv1alpha1.Routing{
		ObjectMeta: metav1.ObjectMeta{Name: "api", Namespace: "team-a"},
		Spec: skiperatorv1alpha1.RoutingSpec{
			Hostname:        "api.EXAMPLE.com",
			RoutingProvider: skiperatorv1alpha1.RoutingProviderStandard,
			Routes:          []skiperatorv1alpha1.Route{{TargetApp: "backend", PathPrefix: "/v1/users", Port: 8080}},
		},
	}

	err := ValidateConflicts(context.Background(), c, routing)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "conflicts with accepted HTTPRoute")
}

func TestRoutingPathConflictWithMatchAllRule(t *testing.T) {
	scheme := runtime.NewScheme()
	resourceschemas.AddSchemas(scheme)
	route := &gatewayapiv1.HTTPRoute{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "accepted",
			Namespace: "team-b",
			Labels: map[string]string{
				"app.kubernetes.io/managed-by":        "skiperator",
				"skiperator.kartverket.no/controller": "routing",
			},
		},
		Spec: gatewayapiv1.HTTPRouteSpec{
			Hostnames: []gatewayapiv1.Hostname{"api.example.com"},
			Rules:     []gatewayapiv1.HTTPRouteRule{{}},
		},
		Status: gatewayapiv1.HTTPRouteStatus{
			RouteStatus: gatewayapiv1.RouteStatus{
				Parents: []gatewayapiv1.RouteParentStatus{
					{Conditions: []metav1.Condition{{Type: string(gatewayapiv1.RouteConditionAccepted), Status: metav1.ConditionTrue}}},
				},
			},
		},
	}
	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(route).Build()
	routing := &skiperatorv1alpha1.Routing{
		ObjectMeta: metav1.ObjectMeta{Name: "api", Namespace: "team-a"},
		Spec: skiperatorv1alpha1.RoutingSpec{
			Hostname:        "api.example.com",
			RoutingProvider: skiperatorv1alpha1.RoutingProviderStandard,
			Routes:          []skiperatorv1alpha1.Route{{TargetApp: "backend", PathPrefix: "/v1", Port: 8080}},
		},
	}

	err := ValidateConflicts(context.Background(), c, routing)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "conflicts with accepted HTTPRoute")
}

func TestRoutingConflictIgnoresRedirectRoute(t *testing.T) {
	scheme := runtime.NewScheme()
	resourceschemas.AddSchemas(scheme)
	pathType := gatewayapiv1.PathMatchPathPrefix
	path := "/"
	route := &gatewayapiv1.HTTPRoute{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "accepted-redirect",
			Namespace: "team-b",
			Labels: map[string]string{
				"app.kubernetes.io/managed-by":        "skiperator",
				"skiperator.kartverket.no/controller": "routing",
			},
		},
		Spec: gatewayapiv1.HTTPRouteSpec{
			Hostnames: []gatewayapiv1.Hostname{"api.example.com"},
			Rules: []gatewayapiv1.HTTPRouteRule{
				{
					Matches: []gatewayapiv1.HTTPRouteMatch{{Path: &gatewayapiv1.HTTPPathMatch{Type: &pathType, Value: &path}}},
					Filters: []gatewayapiv1.HTTPRouteFilter{{Type: gatewayapiv1.HTTPRouteFilterRequestRedirect}},
				},
			},
		},
		Status: gatewayapiv1.HTTPRouteStatus{
			RouteStatus: gatewayapiv1.RouteStatus{
				Parents: []gatewayapiv1.RouteParentStatus{
					{Conditions: []metav1.Condition{{Type: string(gatewayapiv1.RouteConditionAccepted), Status: metav1.ConditionTrue}}},
				},
			},
		},
	}
	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(route).Build()
	routing := gatewayAPIRouting()

	err := ValidateConflicts(context.Background(), c, routing)

	require.NoError(t, err)
}

func TestRoutingStandaloneConflictsWithSharedListenerSet(t *testing.T) {
	scheme := runtime.NewScheme()
	resourceschemas.AddSchemas(scheme)
	listenerSet := readySharedRoutingListenerSet("API.example.COM")
	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(listenerSet).Build()
	routing := gatewayAPIRouting()

	err := ValidateConflicts(context.Background(), c, routing)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "already has an accepted ListenerSet")
}

func TestRoutingSharedOwnershipAllowsSharedListenerSet(t *testing.T) {
	scheme := runtime.NewScheme()
	resourceschemas.AddSchemas(scheme)
	listenerSet := readySharedRoutingListenerSet("api.example.com")
	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(listenerSet).Build()
	routing := gatewayAPIRouting()
	routing.Spec.Ownership = skiperatorv1alpha1.RoutingOwnershipShared

	err := ValidateConflicts(context.Background(), c, routing)

	require.NoError(t, err)
}

func TestRoutingStandardRoutingKeepsLegacyUntilReady(t *testing.T) {
	scheme := runtime.NewScheme()
	resourceschemas.AddSchemas(scheme)
	routing := gatewayAPIRouting()
	legacy := &istionetworkingv1.VirtualService{ObjectMeta: metav1.ObjectMeta{Name: "api-routing-ingress", Namespace: "team-a"}}
	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(routing, legacy).Build()

	state, err := EvaluateRoutingState(context.Background(), c, routing, routing.GetStatus())
	require.NoError(t, err)

	assert.True(t, state.GenerateLegacyRouting)
	assert.False(t, state.Readiness.Ready)
	assert.Contains(t, state.Readiness.Message, "Certificate")
}

func TestRoutingStandardRoutingPrunesLegacyWhenReady(t *testing.T) {
	scheme := runtime.NewScheme()
	resourceschemas.AddSchemas(scheme)
	routing := gatewayAPIRouting()
	certificateName, err := routing.GetCertificateName(mustHost(t, "api.example.com"))
	require.NoError(t, err)
	objects := []client.Object{
		routing,
		&istionetworkingv1.VirtualService{ObjectMeta: metav1.ObjectMeta{Name: "api-routing-ingress", Namespace: "team-a"}},
		readyGateway(IstioGatewayNamespace, ExternalGatewayName),
		readyCertificate("team-a", certificateName),
		tlsSecret("team-a", certificateName),
		readyListenerSet("team-a", ListenerSetName(RoutingResourcePrefix("api"), "api.example.com")),
		readyHTTPRoute("team-a", RoutingResourcePrefix("api")),
		readyHTTPRoute("team-a", RedirectRouteName(RoutingResourcePrefix("api"))),
	}
	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(objects...).Build()

	state, err := EvaluateRoutingState(context.Background(), c, routing, routing.GetStatus())
	require.NoError(t, err)

	assert.False(t, state.GenerateLegacyRouting)
	assert.True(t, state.Readiness.Ready)
}

func TestRoutingStandardRoutingGreenfieldSkipsLegacy(t *testing.T) {
	scheme := runtime.NewScheme()
	resourceschemas.AddSchemas(scheme)
	routing := gatewayAPIRouting()
	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(routing).Build()

	state, err := EvaluateRoutingState(context.Background(), c, routing, routing.GetStatus())
	require.NoError(t, err)

	assert.False(t, state.GenerateLegacyRouting)
	assert.False(t, state.Readiness.Ready)
}

func TestRoutingStandardRoutingCustomCertRequiresSecret(t *testing.T) {
	scheme := runtime.NewScheme()
	resourceschemas.AddSchemas(scheme)
	routing := gatewayAPIRouting()
	routing.Spec.Hostname = "api.example.com+custom-tls"
	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(routing).Build()

	state, err := EvaluateRoutingState(context.Background(), c, routing, routing.GetStatus())
	require.NoError(t, err)

	assert.False(t, state.Readiness.Ready)
	assert.Contains(t, state.Readiness.Message, "Secret team-a/custom-tls")
}

func TestListenerSetReadyWaitsForListenerStatus(t *testing.T) {
	scheme := runtime.NewScheme()
	resourceschemas.AddSchemas(scheme)
	listenerSetName := ListenerSetName("app", "app.example.com")
	listenerSet := &gatewayapiv1.ListenerSet{
		ObjectMeta: metav1.ObjectMeta{Name: listenerSetName, Namespace: "team-a"},
		Spec: gatewayapiv1.ListenerSetSpec{
			ParentRef: gatewayapiv1.ParentGatewayReference{
				Name:      gatewayapiv1.ObjectName(ExternalGatewayName),
				Namespace: gatewayNamespacePtr(IstioGatewayNamespace),
			},
			Listeners: []gatewayapiv1.ListenerEntry{{Name: "http"}, {Name: "https"}},
		},
		Status: gatewayapiv1.ListenerSetStatus{
			Conditions: []metav1.Condition{
				{Type: string(gatewayapiv1.ListenerSetConditionAccepted), Status: metav1.ConditionTrue},
				{Type: string(gatewayapiv1.ListenerSetConditionProgrammed), Status: metav1.ConditionTrue},
			},
		},
	}
	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(listenerSet, readyGateway(IstioGatewayNamespace, ExternalGatewayName), readyTLSSecret("team-a", "tls")).Build()

	ready := standardApplicationReadiness(t, context.Background(), c)

	assert.False(t, ready.Ready)
	assert.Contains(t, ready.Message, "listener status")
}

func TestListenerSetReadyReportsMissingParentGateway(t *testing.T) {
	scheme := runtime.NewScheme()
	resourceschemas.AddSchemas(scheme)
	listenerSet := readyListenerSet("team-a", ListenerSetName("app", "app.example.com"))
	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(listenerSet, readyTLSSecret("team-a", "tls")).Build()

	ready := standardApplicationReadiness(t, context.Background(), c)

	assert.False(t, ready.Ready)
	assert.Contains(t, ready.Message, "parent Gateway istio-gateways/istio-external does not exist")
}

func TestListenerSetReadyReportsUnprogrammedParentGateway(t *testing.T) {
	scheme := runtime.NewScheme()
	resourceschemas.AddSchemas(scheme)
	listenerSet := readyListenerSet("team-a", ListenerSetName("app", "app.example.com"))
	gateway := &gatewayapiv1.Gateway{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ExternalGatewayName,
			Namespace: IstioGatewayNamespace,
		},
	}
	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(listenerSet, gateway, readyTLSSecret("team-a", "tls")).Build()

	ready := standardApplicationReadiness(t, context.Background(), c)

	assert.False(t, ready.Ready)
	assert.Contains(t, ready.Message, "parent Gateway istio-gateways/istio-external is not yet programmed")
}

func gatewayHostname(hostname string) *gatewayapiv1.Hostname {
	h := gatewayapiv1.Hostname(hostname)
	return &h
}

func gatewayNamespacePtr(namespace string) *gatewayapiv1.Namespace {
	n := gatewayapiv1.Namespace(namespace)
	return &n
}

func gatewayAPIApplication() *skiperatorv1alpha1.Application {
	return &skiperatorv1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{Name: "app", Namespace: "team-a"},
		Spec: skiperatorv1alpha1.ApplicationSpec{
			Image:           "image",
			Port:            8080,
			Ingresses:       []string{"app.example.com"},
			RoutingProvider: skiperatorv1alpha1.RoutingProviderStandard,
		},
	}
}

func gatewayAPIRouting() *skiperatorv1alpha1.Routing {
	return &skiperatorv1alpha1.Routing{
		ObjectMeta: metav1.ObjectMeta{Name: "api", Namespace: "team-a"},
		Spec: skiperatorv1alpha1.RoutingSpec{
			Hostname:        "api.example.com",
			RoutingProvider: skiperatorv1alpha1.RoutingProviderStandard,
			Routes: []skiperatorv1alpha1.Route{
				{TargetApp: "backend", PathPrefix: "/", Port: 8080},
			},
		},
	}
}

func standardApplicationReadiness(t *testing.T, ctx context.Context, c client.Client) Readiness {
	t.Helper()
	application := gatewayAPIApplication()
	application.Spec.Ingresses = []string{"app.example.com+tls"}
	state, err := EvaluateRoutingState(ctx, c, application, application.GetStatus())
	require.NoError(t, err)
	return state.Readiness
}

func mustHost(t *testing.T, hostname string) *commontypes.Host {
	t.Helper()
	host, err := commontypes.NewHost(hostname)
	require.NoError(t, err)
	return host
}

func readyCertificate(namespace string, name string) *certmanagerv1.Certificate {
	return &certmanagerv1.Certificate{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
		Status: certmanagerv1.CertificateStatus{
			Conditions: []certmanagerv1.CertificateCondition{
				{Type: certmanagerv1.CertificateConditionReady, Status: certmanagermetav1.ConditionTrue},
			},
		},
	}
}

func tlsSecret(namespace string, name string) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
		Type:       corev1.SecretTypeTLS,
		Data: map[string][]byte{
			corev1.TLSCertKey:       []byte("crt"),
			corev1.TLSPrivateKeyKey: []byte("key"),
		},
	}
}

func readyListenerSet(namespace string, name string) *gatewayapiv1.ListenerSet {
	gatewayNamespace := gatewayapiv1.Namespace(IstioGatewayNamespace)
	return &gatewayapiv1.ListenerSet{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
		Spec: gatewayapiv1.ListenerSetSpec{
			ParentRef: gatewayapiv1.ParentGatewayReference{
				Name:      gatewayapiv1.ObjectName(ExternalGatewayName),
				Namespace: &gatewayNamespace,
			},
			Listeners: []gatewayapiv1.ListenerEntry{{Name: "https"}},
		},
		Status: gatewayapiv1.ListenerSetStatus{
			Conditions: []metav1.Condition{
				{Type: string(gatewayapiv1.ListenerSetConditionAccepted), Status: metav1.ConditionTrue},
				{Type: string(gatewayapiv1.ListenerSetConditionProgrammed), Status: metav1.ConditionTrue},
			},
			Listeners: []gatewayapiv1.ListenerEntryStatus{
				{
					Name: "https",
					Conditions: []metav1.Condition{
						{Type: string(gatewayapiv1.ListenerConditionResolvedRefs), Status: metav1.ConditionTrue},
					},
				},
			},
		},
	}
}

func readySharedRoutingListenerSet(hostname string) *gatewayapiv1.ListenerSet {
	listenerSet := readyListenerSet(IstioGatewayNamespace, SharedListenerSetName(hostname))
	listenerSet.Labels = map[string]string{
		"app.kubernetes.io/managed-by":        "skiperator",
		"skiperator.kartverket.no/controller": "routing-shared",
	}
	listenerSet.Spec.Listeners = []gatewayapiv1.ListenerEntry{{Name: "https", Hostname: gatewayHostname(hostname)}}
	listenerSet.Status.Listeners = []gatewayapiv1.ListenerEntryStatus{
		{
			Name: "https",
			Conditions: []metav1.Condition{
				{Type: string(gatewayapiv1.ListenerConditionResolvedRefs), Status: metav1.ConditionTrue},
			},
		},
	}
	return listenerSet
}

func readyTLSSecret(namespace string, name string) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: name},
		Type:       corev1.SecretTypeTLS,
		Data: map[string][]byte{
			corev1.TLSCertKey:       []byte("cert"),
			corev1.TLSPrivateKeyKey: []byte("key"),
		},
	}
}

func readyGateway(namespace string, name string) *gatewayapiv1.Gateway {
	return &gatewayapiv1.Gateway{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
		Status: gatewayapiv1.GatewayStatus{
			Conditions: []metav1.Condition{
				{Type: string(gatewayapiv1.GatewayConditionProgrammed), Status: metav1.ConditionTrue},
			},
		},
	}
}

func readyHTTPRoute(namespace string, name string) *gatewayapiv1.HTTPRoute {
	return &gatewayapiv1.HTTPRoute{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
		Status: gatewayapiv1.HTTPRouteStatus{
			RouteStatus: gatewayapiv1.RouteStatus{
				Parents: []gatewayapiv1.RouteParentStatus{
					{
						Conditions: []metav1.Condition{
							{Type: string(gatewayapiv1.RouteConditionAccepted), Status: metav1.ConditionTrue},
							{Type: string(gatewayapiv1.RouteConditionResolvedRefs), Status: metav1.ConditionTrue},
						},
					},
				},
			},
		},
	}
}
