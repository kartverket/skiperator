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
	ListenerSetName string
}

type standardRoutingPlan struct {
	namespace       string
	routeBaseName   string
	redirectToHTTPS bool
	hosts           common.HostCollection
	certificateName func(*common.Host) (string, error)
}

func legacyRoutingExists(ctx context.Context, c client.Client, routable legacyRoutable) bool {
	virtualService := &istionetworkingv1.VirtualService{}
	if legacyResourceExists(ctx, c, types.NamespacedName{Namespace: routable.GetNamespace(), Name: routable.GetVirtualServiceName()}, virtualService) {
		return true
	}

	gatewayNames, err := routable.GetGatewayNames()
	if err != nil {
		return false
	}
	for _, name := range gatewayNames {
		gateway := &istionetworkingv1.Gateway{}
		if legacyResourceExists(ctx, c, types.NamespacedName{Namespace: routable.GetNamespace(), Name: name}, gateway) {
			return true
		}
	}
	return false
}

func legacyResourceExists(ctx context.Context, c client.Client, key types.NamespacedName, obj client.Object) bool {
	err := retry.OnError(retry.DefaultRetry, func(err error) bool {
		return !apierrors.IsNotFound(err)
	}, func() error {
		return c.Get(ctx, key, obj)
	})
	return err == nil
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
		routeBaseName:   routing.Name,
		redirectToHTTPS: routing.GetRedirectToHTTPS(),
		hosts:           hosts,
		certificateName: routing.GetCertificateName,
	})
}

func standardRoutingReadiness(ctx context.Context, c client.Client, routable Routable, plan standardRoutingPlan) Readiness {
	if plan.hosts.Count() == 0 {
		return Readiness{Ready: true, Message: "object has no Gateway API hosts"}
	}

	routeNames := []string{RouteName(plan.routeBaseName)}
	if plan.redirectToHTTPS {
		routeNames = append(routeNames, RedirectRouteName(plan.routeBaseName))
	}
	standardHosts, err := standardHostsFor(routable, plan.hosts, plan.certificateName)
	if err != nil {
		return Readiness{Message: err.Error()}
	}
	return standardRoutingReady(ctx, c, plan.namespace, routeNames, standardHosts)
}

func standardHostsFor(routable Routable, hosts common.HostCollection, certificateName func(*common.Host) (string, error)) ([]standardHost, error) {
	standardHosts := make([]standardHost, 0, hosts.Count())
	for _, host := range hosts.AllHosts() {
		name, err := certificateName(host)
		if err != nil {
			return nil, err
		}
		standardHosts = append(standardHosts, standardHost{
			Hostname:        host.Hostname,
			CertificateName: name,
			CustomSecret:    host.CustomCertificateSecret,
			ListenerSetName: ListenerSetName(routable.GetName(), host.Hostname),
		})
	}
	return standardHosts, nil
}

// standardRoutingReady returns the first missing or unready dependency.
//
// This deliberately behaves like an ordered probe rather than a state machine.
// The caller only needs to know if the whole Gateway API path is ready, and if
// not, which dependency blocks safe legacy pruning.
func standardRoutingReady(ctx context.Context, c client.Client, namespace string, routeNames []string, hosts []standardHost) Readiness {
	for _, host := range hosts {
		certificateName := host.CertificateName
		if host.CustomSecret == nil {
			if ready := managedCertificateReady(ctx, c, namespace, certificateName); !ready.Ready {
				return ready
			}
		} else {
			certificateName = *host.CustomSecret
		}
		if ready := tlsSecretReady(ctx, c, namespace, certificateName); !ready.Ready {
			return ready
		}
		if ready := listenerSetReady(ctx, c, namespace, host.ListenerSetName); !ready.Ready {
			return ready
		}
	}
	for _, routeName := range routeNames {
		if ready := httpRouteReady(ctx, c, namespace, routeName); !ready.Ready {
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
