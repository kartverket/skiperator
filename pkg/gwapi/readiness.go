package gwapi

import (
	"context"
	"fmt"

	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	certmanagermetav1 "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
	"github.com/kartverket/skiperator/api/common"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	istionetworkingv1 "istio.io/client-go/pkg/apis/networking/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gatewayapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

// Readiness reports whether the standard Gateway API path is safe to use.
// Message contains the first blocking dependency when Ready is false.
type Readiness struct {
	Ready   bool
	Message string
}

type standardHost struct {
	Hostname        string
	CertificateName string
	CustomSecret    *string
	Namespace       string
	ListenerSetName string
}

type routeCheck struct {
	Namespace string
	Name      string
}

type standardRoutingPlan struct {
	namespace       string
	routeBaseName   string
	redirectToHTTPS bool
	hosts           common.HostCollection
	certificateName func(*common.Host) (string, error)
	sharedRouting   bool
}

// legacyRoutingExists reports whether legacy Istio routing resources are
// present. A non-NotFound API error is returned rather than swallowed: a
// transient error must not be read as "legacy absent", or the migration state
// machine would prune legacy routing while standard routing is not yet ready.
func legacyRoutingExists(ctx context.Context, c client.Client, routable legacyRoutable) (bool, error) {
	virtualService := &istionetworkingv1.VirtualService{}
	exists, err := legacyResourceExists(ctx, c, types.NamespacedName{Namespace: routable.GetNamespace(), Name: routable.GetVirtualServiceName()}, virtualService)
	if err != nil {
		return false, err
	}
	if exists {
		return true, nil
	}

	gatewayNames, err := routable.GetGatewayNames()
	if err != nil {
		return false, err
	}
	for _, name := range gatewayNames {
		gateway := &istionetworkingv1.Gateway{}
		exists, err := legacyResourceExists(ctx, c, types.NamespacedName{Namespace: routable.GetNamespace(), Name: name}, gateway)
		if err != nil {
			return false, err
		}
		if exists {
			return true, nil
		}
	}
	return false, nil
}

func legacyResourceExists(ctx context.Context, c client.Client, key types.NamespacedName, obj client.Object) (bool, error) {
	err := retry.OnError(retry.DefaultRetry, func(err error) bool {
		return !apierrors.IsNotFound(err)
	}, func() error {
		return c.Get(ctx, key, obj)
	})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// applicationStandardRoutingReady checks the Gateway API resources generated
// for one Application.
//
// In legacy Istio terms this replaces "VirtualService and Gateway exist" with
// "certificates are usable, ListenerSets are accepted and programmed, and
// HTTPRoutes are accepted by their parent Gateway".
func applicationStandardRoutingReady(ctx context.Context, c client.Client, application *skiperatorv1alpha1.Application) Readiness {
	hosts, err := application.Hostnames()
	if err != nil {
		return Readiness{Message: err.Error()}
	}

	return standardRoutingReadiness(ctx, c, application, standardRoutingPlan{
		namespace:       application.Namespace,
		routeBaseName:   application.Name,
		redirectToHTTPS: application.Spec.RedirectToHTTPS != nil && *application.Spec.RedirectToHTTPS,
		hosts:           hosts,
		certificateName: application.GetCertificateName,
	})
}

// routingStandardRoutingReady checks the Gateway API resources generated for
// one Routing object. Routing may have several backend rules, but they are
// represented by a single backend HTTPRoute plus an optional redirect route.
func routingStandardRoutingReady(ctx context.Context, c client.Client, routing *skiperatorv1alpha1.Routing) Readiness {
	hosts, err := routing.Hostnames()
	if err != nil {
		return Readiness{Message: err.Error()}
	}
	return standardRoutingReadiness(ctx, c, routing, standardRoutingPlan{
		namespace:       routing.Namespace,
		routeBaseName:   RoutingResourcePrefix(routing.Name),
		redirectToHTTPS: routing.GetRedirectToHTTPS(),
		hosts:           hosts,
		certificateName: routing.GetCertificateName,
		sharedRouting:   routing.UsesSharedOwnership(),
	})
}

func standardRoutingReadiness(ctx context.Context, c client.Client, routable Routable, plan standardRoutingPlan) Readiness {
	if plan.hosts.Count() == 0 {
		return Readiness{Ready: true, Message: "object has no Gateway API hosts"}
	}

	routeChecks := []routeCheck{{Namespace: plan.namespace, Name: plan.routeBaseName}}
	if plan.redirectToHTTPS {
		redirectRoute := routeCheck{Namespace: plan.namespace, Name: RedirectRouteName(plan.routeBaseName)}
		if plan.sharedRouting {
			redirectRoute.Namespace = IstioGatewayNamespace
			redirectRoute.Name = SharedRedirectRouteName(plan.hosts.AllHosts()[0].Hostname)
		}
		routeChecks = append(routeChecks, redirectRoute)
	}
	standardHosts, err := standardHostsFor(routable, plan)
	if err != nil {
		return Readiness{Message: err.Error()}
	}
	return standardRoutingReady(ctx, c, routeChecks, standardHosts)
}

func standardHostsFor(routable Routable, plan standardRoutingPlan) ([]standardHost, error) {
	standardHosts := make([]standardHost, 0, plan.hosts.Count())
	for _, host := range plan.hosts.AllHosts() {
		name, err := plan.certificateName(host)
		if err != nil {
			return nil, err
		}
		namespace := routable.GetNamespace()
		// Key off the plan's (kind-qualified) base name, the same prefix the
		// generator uses, so Application and Routing ListenerSets stay distinct.
		listenerSetName := ListenerSetName(plan.routeBaseName, host.Hostname)
		if plan.sharedRouting {
			namespace = IstioGatewayNamespace
			listenerSetName = SharedListenerSetName(host.Hostname)
		}
		standardHosts = append(standardHosts, standardHost{
			Hostname:        host.Hostname,
			CertificateName: name,
			CustomSecret:    host.CustomCertificateSecret,
			Namespace:       namespace,
			ListenerSetName: listenerSetName,
		})
	}
	return standardHosts, nil
}

// standardRoutingReady returns the first missing or unready dependency.
//
// This deliberately behaves like an ordered probe rather than a state machine.
// The caller only needs to know if the whole Gateway API path is ready, and if
// not, which dependency blocks safe legacy pruning.
func standardRoutingReady(ctx context.Context, c client.Client, routeChecks []routeCheck, hosts []standardHost) Readiness {
	for _, host := range hosts {
		certificateName := host.CertificateName
		if host.CustomSecret == nil {
			if ready := managedCertificateReady(ctx, c, host.Namespace, certificateName); !ready.Ready {
				return ready
			}
		} else {
			certificateName = *host.CustomSecret
		}
		if ready := tlsSecretReady(ctx, c, host.Namespace, certificateName); !ready.Ready {
			return ready
		}
		if ready := listenerSetReady(ctx, c, host.Namespace, host.ListenerSetName); !ready.Ready {
			return ready
		}
	}
	for _, route := range routeChecks {
		if ready := httpRouteReady(ctx, c, route.Namespace, route.Name); !ready.Ready {
			return ready
		}
	}
	return Readiness{Ready: true, Message: "Gateway API routing is ready"}
}

func managedCertificateReady(ctx context.Context, c client.Client, namespace string, name string) Readiness {
	certificate := &certmanagerv1.Certificate{}
	if err := c.Get(ctx, types.NamespacedName{Namespace: namespace, Name: name}, certificate); err != nil {
		if apierrors.IsNotFound(err) {
			return Readiness{Message: fmt.Sprintf("waiting for Certificate %s/%s", namespace, name)}
		}
		return Readiness{Message: err.Error()}
	}
	for _, condition := range certificate.Status.Conditions {
		if condition.Type == certmanagerv1.CertificateConditionReady && condition.Status == certmanagermetav1.ConditionTrue {
			return Readiness{Ready: true, Message: fmt.Sprintf("Certificate %s/%s is ready", namespace, name)}
		}
	}
	return Readiness{Message: fmt.Sprintf("waiting for Certificate %s/%s Ready=True", namespace, name)}
}

func tlsSecretReady(ctx context.Context, c client.Client, namespace string, name string) Readiness {
	secret := &corev1.Secret{}
	if err := c.Get(ctx, types.NamespacedName{Namespace: namespace, Name: name}, secret); err != nil {
		if apierrors.IsNotFound(err) {
			return Readiness{Message: fmt.Sprintf("waiting for Secret %s/%s", namespace, name)}
		}
		return Readiness{Message: err.Error()}
	}
	if secret.Type != corev1.SecretTypeTLS {
		return Readiness{Message: fmt.Sprintf("waiting for Secret %s/%s to be kubernetes.io/tls", namespace, name)}
	}
	if len(secret.Data[corev1.TLSCertKey]) == 0 || len(secret.Data[corev1.TLSPrivateKeyKey]) == 0 {
		return Readiness{Message: fmt.Sprintf("waiting for Secret %s/%s tls.crt and tls.key", namespace, name)}
	}
	return Readiness{Ready: true, Message: fmt.Sprintf("Secret %s/%s is ready", namespace, name)}
}

func listenerSetReady(ctx context.Context, c client.Client, namespace string, name string) Readiness {
	listenerSet := &gatewayapiv1.ListenerSet{}
	if err := c.Get(ctx, types.NamespacedName{Namespace: namespace, Name: name}, listenerSet); err != nil {
		if apierrors.IsNotFound(err) {
			return Readiness{Message: fmt.Sprintf("waiting for ListenerSet %s/%s", namespace, name)}
		}
		return Readiness{Message: err.Error()}
	}
	if ready := parentGatewayReady(ctx, c, listenerSet); !ready.Ready {
		return ready
	}
	if !meta.IsStatusConditionTrue(listenerSet.Status.Conditions, string(gatewayapiv1.ListenerSetConditionAccepted)) {
		return Readiness{Message: fmt.Sprintf("waiting for ListenerSet %s/%s Accepted=True", namespace, name)}
	}
	if !meta.IsStatusConditionTrue(listenerSet.Status.Conditions, string(gatewayapiv1.ListenerSetConditionProgrammed)) {
		return Readiness{Message: fmt.Sprintf("waiting for ListenerSet %s/%s Programmed=True", namespace, name)}
	}
	if len(listenerSet.Status.Listeners) < len(listenerSet.Spec.Listeners) {
		return Readiness{Message: fmt.Sprintf("waiting for ListenerSet %s/%s listener status", namespace, name)}
	}
	listenerStatusByName := make(map[gatewayapiv1.SectionName]gatewayapiv1.ListenerEntryStatus, len(listenerSet.Status.Listeners))
	for _, listener := range listenerSet.Status.Listeners {
		listenerStatusByName[listener.Name] = listener
	}
	for _, specListener := range listenerSet.Spec.Listeners {
		listener, ok := listenerStatusByName[specListener.Name]
		if !ok {
			return Readiness{Message: fmt.Sprintf("waiting for ListenerSet %s/%s listener %s status", namespace, name, specListener.Name)}
		}
		if meta.IsStatusConditionTrue(listener.Conditions, string(gatewayapiv1.ListenerEntryConditionConflicted)) {
			return Readiness{Message: fmt.Sprintf("waiting for ListenerSet %s/%s listener %s Conflicted=False", namespace, name, listener.Name)}
		}
		if !meta.IsStatusConditionTrue(listener.Conditions, string(gatewayapiv1.ListenerConditionResolvedRefs)) {
			return Readiness{Message: fmt.Sprintf("waiting for ListenerSet %s/%s listener %s ResolvedRefs=True", namespace, name, listener.Name)}
		}
	}
	return Readiness{Ready: true, Message: fmt.Sprintf("ListenerSet %s/%s is ready", namespace, name)}
}

func parentGatewayReady(ctx context.Context, c client.Client, listenerSet *gatewayapiv1.ListenerSet) Readiness {
	namespace := listenerSet.Namespace
	if listenerSet.Spec.ParentRef.Namespace != nil {
		namespace = string(*listenerSet.Spec.ParentRef.Namespace)
	}
	name := string(listenerSet.Spec.ParentRef.Name)

	gateway := &gatewayapiv1.Gateway{}
	if err := c.Get(ctx, types.NamespacedName{Namespace: namespace, Name: name}, gateway); err != nil {
		if apierrors.IsNotFound(err) {
			return Readiness{Message: fmt.Sprintf("parent Gateway %s/%s does not exist", namespace, name)}
		}
		return Readiness{Message: err.Error()}
	}
	if !meta.IsStatusConditionTrue(gateway.Status.Conditions, string(gatewayapiv1.GatewayConditionProgrammed)) {
		return Readiness{Message: fmt.Sprintf("parent Gateway %s/%s is not yet programmed", namespace, name)}
	}
	return Readiness{Ready: true, Message: fmt.Sprintf("parent Gateway %s/%s is ready", namespace, name)}
}

func httpRouteReady(ctx context.Context, c client.Client, namespace string, name string) Readiness {
	route := &gatewayapiv1.HTTPRoute{}
	if err := c.Get(ctx, types.NamespacedName{Namespace: namespace, Name: name}, route); err != nil {
		if apierrors.IsNotFound(err) {
			return Readiness{Message: fmt.Sprintf("waiting for HTTPRoute %s/%s", namespace, name)}
		}
		return Readiness{Message: err.Error()}
	}
	if len(route.Status.Parents) == 0 {
		return Readiness{Message: fmt.Sprintf("waiting for HTTPRoute %s/%s parent status", namespace, name)}
	}
	for _, parent := range route.Status.Parents {
		if !meta.IsStatusConditionTrue(parent.Conditions, string(gatewayapiv1.RouteConditionAccepted)) {
			return Readiness{Message: fmt.Sprintf("waiting for HTTPRoute %s/%s Accepted=True", namespace, name)}
		}
		if !meta.IsStatusConditionTrue(parent.Conditions, string(gatewayapiv1.RouteConditionResolvedRefs)) {
			return Readiness{Message: fmt.Sprintf("waiting for HTTPRoute %s/%s ResolvedRefs=True", namespace, name)}
		}
	}
	return Readiness{Ready: true, Message: fmt.Sprintf("HTTPRoute %s/%s is ready", namespace, name)}
}
