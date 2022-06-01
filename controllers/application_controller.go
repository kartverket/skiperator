/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"

	istioApiNetworkingv1beta1 "istio.io/api/networking/v1beta1"
	istioApiSecurityv1beta1 "istio.io/api/security/v1beta1"
	istioNetworkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	istioSecurityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
)

// ApplicationReconciler reconciles a Application object
type ApplicationReconciler struct {
	client.Client
	Scheme             *runtime.Scheme
	PrevServiceEntries []string
}

//+kubebuilder:rbac:groups=skiperator.kartverket.no,resources=applications,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=skiperator.kartverket.no,resources=applications/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=skiperator.kartverket.no,resources=applications/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=replicasets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=autoscaling,resources=horizontalpodautoscalers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=networking.k8s.io,resources=networkpolicies,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=networking.istio.io,resources=gateways,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=networking.istio.io,resources=virtualservices,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=networking.istio.io,resources=serviceentries,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=networking.istio.io,resources=sidecars,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=security.istio.io,resources=peerauthentications,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (reconciler *ApplicationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// Lookup the Application instance for this reconcile request
	app := &skiperatorv1alpha1.Application{}
	err := reconciler.Get(ctx, req.NamespacedName, app)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info("Application resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get Application")
		return ctrl.Result{}, err
	}

	// Ensure status fields are initialized
	if app.Status.OperationResults == nil {
		app.Status.OperationResults = map[string]controllerutil.OperationResult{}
		reconciler.Status().Update(ctx, app)
	}
	if app.Status.Errors == nil {
		app.Status.Errors = map[string]string{}
		reconciler.Status().Update(ctx, app)
	}

	log.Info("The incoming application object is", "Application", app)

	// Deployment: Check if already exists, if not create a new one
	deployment := &appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: app.Name, Namespace: app.Namespace}}
	reconciler.reconcileObject("Deployment", ctx, app, deployment, func() {
		reconciler.addDeploymentData(ctx, app, deployment)
	})

	// HorizontalPodAutoscaler: Check if already exists, if not create a new one
	if app.Spec.Replicas != nil && !app.Spec.Replicas.DisableAutoScaling {
		autoscaler := &autoscalingv1.HorizontalPodAutoscaler{ObjectMeta: metav1.ObjectMeta{Name: app.Name, Namespace: app.Namespace}}
		reconciler.reconcileObject("HorizontalPodAutoscaler", ctx, app, autoscaler, func() {
			reconciler.addAutoscalerData(app, autoscaler)
		})
	} else if app.Spec.Replicas == nil || app.Spec.Replicas.DisableAutoScaling {
		autoscaler := &autoscalingv1.HorizontalPodAutoscaler{ObjectMeta: metav1.ObjectMeta{Name: app.Name, Namespace: app.Namespace}}
		reconciler.Delete(ctx, autoscaler)
	}

	// Service: Check if already exists, if not create a new one
	service := &v1.Service{ObjectMeta: metav1.ObjectMeta{Name: app.Name, Namespace: app.Namespace}}
	reconciler.reconcileObject("Service", ctx, app, service, func() {
		reconciler.addServiceData(app, service)
	})

	// Gateway ingress: Check if already exists, if not create a new one
	if len(app.Spec.Ingresses) > 0 {
		gateway := &istioNetworkingv1beta1.Gateway{ObjectMeta: metav1.ObjectMeta{Name: app.Name + "-ingress", Namespace: app.Namespace}}
		reconciler.reconcileObject("Gateway", ctx, app, gateway, func() {
			reconciler.addIngressGatewayData(app, gateway)
		})
	} else {
		gateway := &istioNetworkingv1beta1.Gateway{ObjectMeta: metav1.ObjectMeta{Name: app.Name + "-ingress", Namespace: app.Namespace}}
		reconciler.Delete(ctx, gateway)
	}

	// Gateway egress: Check if already exists, if not create a new one
	if app.Spec.AccessPolicy != nil && app.Spec.AccessPolicy.Outbound != nil && len(app.Spec.AccessPolicy.Outbound.External) > 0 {
		gateway := &istioNetworkingv1beta1.Gateway{ObjectMeta: metav1.ObjectMeta{Name: app.Name + "-egress", Namespace: app.Namespace}}
		reconciler.reconcileObject("Gateway", ctx, app, gateway, func() {
			reconciler.addEgressGatewayData(app, gateway)
		})
	} else {
		gateway := &istioNetworkingv1beta1.Gateway{ObjectMeta: metav1.ObjectMeta{Name: app.Name + "-egress", Namespace: app.Namespace}}
		reconciler.Delete(ctx, gateway)
	}

	// ServiceEntry: Create all that are defined in outgoing external
	if app.Spec.AccessPolicy != nil && app.Spec.AccessPolicy.Outbound != nil && app.Spec.AccessPolicy.Outbound.External != nil {
		for _, rule := range app.Spec.AccessPolicy.Outbound.External {
			serviceEntry := &istioNetworkingv1beta1.ServiceEntry{ObjectMeta: metav1.ObjectMeta{Name: app.Name + "-" + rule.Host, Namespace: app.Namespace}}
			reconciler.reconcileObject("ServiceEntry", ctx, app, serviceEntry, func() {
				reconciler.addServiceEntryData(app, &rule, serviceEntry)
			})
		}
	}

	// Prune dangling ServiceEntries if any get deleted from Application
	for _, entry := range reconciler.PrevServiceEntries {
		// Check if entry is still in use by comparing with previous reconcile
		isUsed := false
		if app.Spec.AccessPolicy != nil && app.Spec.AccessPolicy.Outbound != nil && app.Spec.AccessPolicy.Outbound.External != nil {
			for _, rule := range app.Spec.AccessPolicy.Outbound.External {
				if entry == rule.Host {
					isUsed = true
				}
			}
		}

		// Delete unused entries
		if !isUsed {
			name := app.Name + "-" + entry
			log.Info("Deleting ServiceEntry " + name)
			serviceEntry := &istioNetworkingv1beta1.ServiceEntry{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: app.Namespace}}
			reconciler.Delete(ctx, serviceEntry)
		}
	}

	// Set current ServiceEntries for next reconcile loop
	newServiceEntries := []string{}
	if app.Spec.AccessPolicy != nil && app.Spec.AccessPolicy.Outbound != nil && app.Spec.AccessPolicy.Outbound.External != nil {
		for _, rule := range app.Spec.AccessPolicy.Outbound.External {
			newServiceEntries = append(newServiceEntries, rule.Host)
		}
	}
	reconciler.PrevServiceEntries = newServiceEntries

	// VirtualService Ingress: Check if already exists, if not create a new one
	if len(app.Spec.Ingresses) > 0 {
		virtualServiceIngress := &istioNetworkingv1beta1.VirtualService{ObjectMeta: metav1.ObjectMeta{Name: app.Name + "-ingress", Namespace: app.Namespace}}
		reconciler.reconcileObject("VirtualService", ctx, app, virtualServiceIngress, func() {
			reconciler.addIngressVirtualServiceData(app, virtualServiceIngress)
		})
	} else {
		virtualServiceIngress := &istioNetworkingv1beta1.VirtualService{ObjectMeta: metav1.ObjectMeta{Name: app.Name + "-ingress", Namespace: app.Namespace}}
		reconciler.Delete(ctx, virtualServiceIngress)
	}

	// VirtualService Egress: Check if already exists, if not create a new one
	if app.Spec.AccessPolicy != nil && app.Spec.AccessPolicy.Outbound != nil && len(app.Spec.AccessPolicy.Outbound.External) > 0 {
		virtualServiceEgress := &istioNetworkingv1beta1.VirtualService{ObjectMeta: metav1.ObjectMeta{Name: app.Name + "-egress", Namespace: app.Namespace}}
		reconciler.reconcileObject("VirtualService", ctx, app, virtualServiceEgress, func() {
			reconciler.addEgressVirtualServiceData(app, virtualServiceEgress)
		})
	} else {
		virtualServiceEgress := &istioNetworkingv1beta1.VirtualService{ObjectMeta: metav1.ObjectMeta{Name: app.Name + "-egress", Namespace: app.Namespace}}
		reconciler.Delete(ctx, virtualServiceEgress)
	}

	// NetworkPolicy ingress: Check if already exists, if not create a new one
	networkPolicyIngress := &networkingv1.NetworkPolicy{ObjectMeta: metav1.ObjectMeta{Name: app.Name + "-ingress", Namespace: app.Namespace}}
	reconciler.reconcileObject("NetworkPolicy", ctx, app, networkPolicyIngress, func() {
		reconciler.addIngressNetworkPolicyData(app, networkPolicyIngress)
	})

	// NetworkPolicy egress: Check if already exists, if not create a new one
	networkPolicyEgress := &networkingv1.NetworkPolicy{ObjectMeta: metav1.ObjectMeta{Name: app.Name + "-egress", Namespace: app.Namespace}}
	reconciler.reconcileObject("NetworkPolicy", ctx, app, networkPolicyEgress, func() {
		reconciler.addEgressNetworkPolicyData(app, networkPolicyEgress)
	})

	// PeerAuthentication: Check if already exists, if not create a new one
	/*
		TODO: Make sure traffic between pods works
		peerAuthentication := &istioSecurityv1beta1.PeerAuthentication{ObjectMeta: metav1.ObjectMeta{Name: app.Name, Namespace: app.Namespace}}
		reconciler.reconcileObject("PeerAuthentication", ctx, app, peerAuthentication, func() {
			reconciler.addPeerAuthenticationData(app, peerAuthentication)
		})
	*/

	// Sidecar: Check if already exists, if not create a new one
	/*
		TODO: Make sure traffic between pods works
		sidecar := &istioNetworkingv1beta1.Sidecar{ObjectMeta: metav1.ObjectMeta{Name: app.Name, Namespace: app.Namespace}}
		reconciler.reconcileObject("Sidecar", ctx, app, sidecar, func() {
			reconciler.addSidecarDara(app, sidecar)
		})
	*/

	// TODO make image pull Secret

	return ctrl.Result{}, err
}

func (reconciler *ApplicationReconciler) reconcileObject(ident string, ctx context.Context, app *skiperatorv1alpha1.Application, object client.Object, f func()) {
	log := log.FromContext(ctx)
	statusIdent := ident + "/" + object.GetName()
	op, err := ctrl.CreateOrUpdate(ctx, reconciler.Client, object, func() error {
		// Set object to expected state in memory. If it's different than what
		// the calling `CreateOrUpdate` function gets from Kubernetes, it will send an
		// update request back to the apiserver with the expected state determined here.
		f()

		// Setting controller as owner makes the object garbage collected when Application gets deleted in k8s
		if err := ctrl.SetControllerReference(app, object, reconciler.Scheme); err != nil {
			log.Error(err, "Failed to set owner reference on "+ident, ident+".Namespace", object.GetNamespace(), ident+".Name", object.GetName())
			return err
		}
		return nil

	})
	if err != nil {
		log.Error(err, "Failed to reconcile "+ident, ident+".Namespace", object.GetNamespace(), ident+".Name", object.GetName())
		app.Status.Errors[statusIdent] = err.Error()
	} else {
		log.Info(ident+" reconciled", ident+".Namespace", object.GetNamespace(), ident+".Name", object.GetName(), "Operation", op)
		app.Status.Errors[statusIdent] = "Success"
	}

	app.Status.OperationResults[statusIdent] = op
	reconciler.Status().Update(ctx, app)
}

func (reconciler *ApplicationReconciler) addEgressVirtualServiceData(app *skiperatorv1alpha1.Application, virtualService *istioNetworkingv1beta1.VirtualService) {
	operatorOwnedGatewayName := app.Name + "-egress"
	hasRules := app.Spec.AccessPolicy != nil && app.Spec.AccessPolicy.Outbound != nil && len(app.Spec.AccessPolicy.Outbound.External) > 0
	size := 0
	if hasRules {
		size = size + len(app.Spec.AccessPolicy.Outbound.External)
	}

	if len(virtualService.Spec.Http) < 1 {
		virtualService.Spec.Http = make([]*istioApiNetworkingv1beta1.HTTPRoute, 1)
		virtualService.Spec.Http[0] = &istioApiNetworkingv1beta1.HTTPRoute{
			Match: []*istioApiNetworkingv1beta1.HTTPMatchRequest{{
				Gateways: []string{"mesh"},
			}},
			Route: []*istioApiNetworkingv1beta1.HTTPRouteDestination{{
				Destination: &istioApiNetworkingv1beta1.Destination{
					Host: "egress-external.istio-system.svc.cluster.local",
					Port: &istioApiNetworkingv1beta1.PortSelector{
						Number: 80,
					},
				},
			}},
		}
	}

	// Get all hosts using TLS protocol
	sniHosts := []string{}
	if hasRules {
		for _, host := range app.Spec.AccessPolicy.Outbound.External {
			for _, port := range reconciler.getHostPorts(host) {
				if port.Protocol == "HTTPS" {
					sniHosts = append(sniHosts, host.Host)
				}
			}
		}
	}

	// Initialize TLS when TLS is in use
	if len(virtualService.Spec.Tls) < 1 && len(sniHosts) > 0 {
		virtualService.Spec.Tls = make([]*istioApiNetworkingv1beta1.TLSRoute, 1)
		virtualService.Spec.Tls[0] = &istioApiNetworkingv1beta1.TLSRoute{
			Match: []*istioApiNetworkingv1beta1.TLSMatchAttributes{{
				Gateways: []string{"mesh"},
				SniHosts: sniHosts,
			}},
			Route: []*istioApiNetworkingv1beta1.RouteDestination{{
				Destination: &istioApiNetworkingv1beta1.Destination{
					Host: "egress-external.istio-system.svc.cluster.local",
					Port: &istioApiNetworkingv1beta1.PortSelector{
						Number: 80,
					},
				},
			}},
		}
	} else if len(sniHosts) == 0 {
		// Remove TLS when not in use
		virtualService.Spec.Tls = nil
	}

	if len(virtualService.Spec.Tcp) < 1 {
		virtualService.Spec.Tcp = make([]*istioApiNetworkingv1beta1.TCPRoute, 1)
		virtualService.Spec.Tcp[0] = &istioApiNetworkingv1beta1.TCPRoute{
			Match: []*istioApiNetworkingv1beta1.L4MatchAttributes{{
				Gateways: []string{"mesh"},
			}},
			Route: []*istioApiNetworkingv1beta1.RouteDestination{{
				Destination: &istioApiNetworkingv1beta1.Destination{
					Host: "egress-external.istio-system.svc.cluster.local",
					Port: &istioApiNetworkingv1beta1.PortSelector{
						Number: 80,
					},
				},
			}},
		}
	}

	hosts := make([]string, size)
	if hasRules {
		// Counters for array indexing for each type
		httpi, tlsi, tcpi := 0, 0, 0
		for i, host := range app.Spec.AccessPolicy.Outbound.External {
			hosts[i] = host.Host
			// Set default ports when user does not specify them
			ports := reconciler.getHostPorts(host)

			for _, port := range ports {
				if port.Protocol == "HTTP" || port.Protocol == "HTTP2" {
					if len(virtualService.Spec.Http) < httpi+2 {
						virtualService.Spec.Http = append(virtualService.Spec.Http, &istioApiNetworkingv1beta1.HTTPRoute{})
					}

					http := virtualService.Spec.Http[httpi+1]

					if http.Match == nil {
						http.Match = []*istioApiNetworkingv1beta1.HTTPMatchRequest{{
							Gateways: []string{operatorOwnedGatewayName},
						}}
					}

					if http.Route == nil {
						http.Route = []*istioApiNetworkingv1beta1.HTTPRouteDestination{{
							Destination: &istioApiNetworkingv1beta1.Destination{
								Port: &istioApiNetworkingv1beta1.PortSelector{},
							},
						}}
					}
					http.Route[0].Destination.Host = host.Host
					http.Route[0].Destination.Port.Number = uint32(port.Port)
					virtualService.Spec.Http[httpi+1] = http
					httpi++
				} else if port.Protocol == "HTTPS" {
					if len(virtualService.Spec.Tls) < tlsi+2 {
						virtualService.Spec.Tls = append(virtualService.Spec.Tls, &istioApiNetworkingv1beta1.TLSRoute{})
					}

					tls := virtualService.Spec.Tls[tlsi+1]

					if tls.Match == nil {
						tls.Match = []*istioApiNetworkingv1beta1.TLSMatchAttributes{{
							Gateways: []string{operatorOwnedGatewayName},
						}}
					}

					if tls.Route == nil {
						tls.Route = []*istioApiNetworkingv1beta1.RouteDestination{{
							Destination: &istioApiNetworkingv1beta1.Destination{
								Port: &istioApiNetworkingv1beta1.PortSelector{},
							},
						}}
					}
					tls.Match[0].SniHosts = []string{host.Host}
					tls.Route[0].Destination.Host = host.Host
					tls.Route[0].Destination.Port.Number = uint32(port.Port)
					virtualService.Spec.Tls[tlsi+1] = tls
					tlsi++
				} else if port.Protocol == "TCP" {
					if len(virtualService.Spec.Tcp) < tcpi+2 {
						virtualService.Spec.Tcp = append(virtualService.Spec.Tcp, &istioApiNetworkingv1beta1.TCPRoute{})
					}
					tcp := virtualService.Spec.Tcp[tcpi+1]

					if tcp.Match == nil {
						tcp.Match = []*istioApiNetworkingv1beta1.L4MatchAttributes{{
							Gateways: []string{operatorOwnedGatewayName},
						}}
					}

					if tcp.Route == nil {
						tcp.Route = []*istioApiNetworkingv1beta1.RouteDestination{{
							Destination: &istioApiNetworkingv1beta1.Destination{
								Port: &istioApiNetworkingv1beta1.PortSelector{},
							},
						}}
					}
					tcp.Route[0].Destination.Host = host.Host
					tcp.Route[0].Destination.Port.Number = uint32(port.Port)
					virtualService.Spec.Tcp[tcpi+1] = tcp
					tcpi++
				}
			}
		}
	}

	virtualService.Spec.ExportTo = []string{".", "istio-system"}
	virtualService.Spec.Gateways = []string{"mesh", operatorOwnedGatewayName}
	virtualService.Spec.Hosts = hosts
}

func (*ApplicationReconciler) getHostPorts(host skiperatorv1alpha1.ExternalRule) []skiperatorv1alpha1.Port {
	ports := host.Ports

	if len(ports) == 0 {
		ports = []skiperatorv1alpha1.Port{{
			Name:     "HTTP",
			Protocol: "HTTP",
			Port:     80,
		}, {
			Name:     "HTTPS",
			Protocol: "HTTPS",
			Port:     443,
		}}
	}

	return ports
}

func (reconciler *ApplicationReconciler) addIngressVirtualServiceData(app *skiperatorv1alpha1.Application, virtualService *istioNetworkingv1beta1.VirtualService) {
	if len(virtualService.Spec.Http) < 1 {
		virtualService.Spec.Http = []*istioApiNetworkingv1beta1.HTTPRoute{{
			// TODO: Can we safely omit this when adding TLS?
			// "Port specifies the ports on the host that is being addressed. Many services only expose a single port
			// or label ports with the protocols they support, in these cases it is not required to explicitly select the port."
			// Match: []*istioApiNetworkingv1beta1.HTTPMatchRequest{{
			// 	Port: uint32(app.Spec.Port),
			// }},
			Route: []*istioApiNetworkingv1beta1.HTTPRouteDestination{{
				Destination: &istioApiNetworkingv1beta1.Destination{},
			}},
		}}
	}

	virtualService.Spec.Hosts = app.Spec.Ingresses
	virtualService.Spec.Gateways = []string{app.Name}
	virtualService.Spec.Http[0].Route[0].Destination.Host = app.Name
}

func (reconciler *ApplicationReconciler) addServiceEntryData(app *skiperatorv1alpha1.Application, rule *skiperatorv1alpha1.ExternalRule, serviceEntry *istioNetworkingv1beta1.ServiceEntry) {
	if len(serviceEntry.Spec.Ports) != len(rule.Ports) {
		serviceEntry.Spec.Ports = make([]*istioApiNetworkingv1beta1.Port, len(rule.Ports))
	}

	for i, port := range rule.Ports {
		if serviceEntry.Spec.Ports[i] == nil {
			serviceEntry.Spec.Ports[i] = &istioApiNetworkingv1beta1.Port{}
		}
		serviceEntry.Spec.Ports[i].Name = port.Name
		serviceEntry.Spec.Ports[i].Number = uint32(port.Port)
		serviceEntry.Spec.Ports[i].Protocol = port.Protocol
	}

	serviceEntry.Spec.Hosts = []string{rule.Host}
	serviceEntry.Spec.ExportTo = []string{".", "istio-system"}
	serviceEntry.Spec.Resolution = istioApiNetworkingv1beta1.ServiceEntry_DNS
}

func (reconciler *ApplicationReconciler) addEgressGatewayData(app *skiperatorv1alpha1.Application, gateway *istioNetworkingv1beta1.Gateway) {
	gateway.Spec.Selector = map[string]string{
		"egress": "external",
	}

	count := 0
	for _, host := range app.Spec.AccessPolicy.Outbound.External {
		for range reconciler.getHostPorts(host) {
			count = count + 1
		}
	}

	if len(gateway.Spec.Servers) != len(app.Spec.AccessPolicy.Outbound.External) {
		gateway.Spec.Servers = make([]*istioApiNetworkingv1beta1.Server, count)
	}

	i := 0
	for _, host := range app.Spec.AccessPolicy.Outbound.External {
		for _, port := range reconciler.getHostPorts(host) {
			if gateway.Spec.Servers[i] == nil {
				gateway.Spec.Servers[i] = &istioApiNetworkingv1beta1.Server{
					Port: &istioApiNetworkingv1beta1.Port{},
				}
			}

			gateway.Spec.Servers[i].Port.Number = uint32(port.Port)
			gateway.Spec.Servers[i].Port.Name = port.Name
			gateway.Spec.Servers[i].Port.Protocol = port.Protocol
			gateway.Spec.Servers[i].Hosts = []string{host.Host}

			if port.Protocol == "HTTPS" {
				if gateway.Spec.Servers[i].Tls == nil {
					gateway.Spec.Servers[i].Tls = &istioApiNetworkingv1beta1.ServerTLSSettings{}
				}
				// TODO the value below is omitted when viewed in k8s due to JSON
				// omitonly on the Tls.Mode struct. Bug in istio API?
				gateway.Spec.Servers[i].Tls.Mode = istioApiNetworkingv1beta1.ServerTLSSettings_PASSTHROUGH
			}

			i = i + 1
		}
	}
}

func (reconciler *ApplicationReconciler) addIngressGatewayData(app *skiperatorv1alpha1.Application, gateway *istioNetworkingv1beta1.Gateway) {
	gateway.Spec.Selector = map[string]string{
		"istio": "ingressgateway",
	}

	if len(gateway.Spec.Servers) == 0 {
		gateway.Spec.Servers = []*istioApiNetworkingv1beta1.Server{{
			Port: &istioApiNetworkingv1beta1.Port{},
		}}
	}

	gateway.Spec.Servers[0].Port.Number = 80
	gateway.Spec.Servers[0].Port.Name = "HTTP"
	gateway.Spec.Servers[0].Port.Protocol = "HTTP"
	gateway.Spec.Servers[0].Hosts = app.Spec.Ingresses

	/* TODO: Add HTTPS routes.
	It fails in validation when applied due to "configuration is invalid: server must have TLS settings for HTTPS/TLS protocols"
	gateway.Spec.Servers[0].Port.Number = 443
	gateway.Spec.Servers[0].Port.Name = "HTTPS"
	gateway.Spec.Servers[0].Port.Protocol = "HTTPS"
	gateway.Spec.Servers[0].Hosts = app.Spec.Ingresses
	*/
}

func (reconciler *ApplicationReconciler) addServiceData(app *skiperatorv1alpha1.Application, service *v1.Service) {
	labels := labelsForApplication(app)

	if len(service.Spec.Ports) < 1 {
		service.Spec.Ports = make([]v1.ServicePort, 1)
	}

	service.Spec.Selector = labels
	service.Spec.Type = v1.ServiceTypeClusterIP
	service.Spec.Ports[0].Port = int32(app.Spec.Port)
	service.Spec.Ports[0].TargetPort = intstr.FromInt(app.Spec.Port)

	tcp := "tcp"
	http := "http"

	if app.Spec.Port == 5432 {
		service.Spec.Ports[0].AppProtocol = &tcp
		service.Spec.Ports[0].Name = "tcp"
	} else {
		service.Spec.Ports[0].AppProtocol = &http
		service.Spec.Ports[0].Name = "http"
	}
}

func (reconciler *ApplicationReconciler) addAutoscalerData(app *skiperatorv1alpha1.Application, autoscaler *autoscalingv1.HorizontalPodAutoscaler) {
	autoscaler.Spec.ScaleTargetRef.APIVersion = "apps/v1"
	autoscaler.Spec.ScaleTargetRef.Kind = "Deployment"
	autoscaler.Spec.ScaleTargetRef.Name = app.Name
	autoscaler.Spec.MinReplicas = app.Spec.Replicas.Min
	autoscaler.Spec.MaxReplicas = app.Spec.Replicas.Max
	autoscaler.Spec.TargetCPUUtilizationPercentage = app.Spec.Replicas.CpuThresholdPercentage
}

func (reconciler *ApplicationReconciler) addDeploymentData(ctx context.Context, app *skiperatorv1alpha1.Application, deployment *appsv1.Deployment) {
	labels := labelsForApplication(app)
	var uid int64 = 150
	yes := true
	no := false
	var replicas int32 = 1
	if app.Spec.Replicas != nil && app.Spec.Replicas.Min != nil {
		replicas = *app.Spec.Replicas.Min
	}

	if deployment.Spec.Selector == nil {
		// This block only runs on initial creation
		// It holds initialization of various objects in addition to immutable objects
		// which are not possible to edit after creation
		volumeMounts, volumes := reconciler.buildVolumes(app)

		deployment.Spec.Selector = &metav1.LabelSelector{
			MatchLabels: labels,
		}
		deployment.Spec.Template.ObjectMeta.Annotations = map[string]string{
			"prometheus.io/scrape": "true",
		}
		deployment.Spec.Template.Spec.Containers = make([]v1.Container, 1)
		deployment.Spec.Template.Spec.Containers[0].Ports = make([]v1.ContainerPort, 1)
		deployment.Spec.Template.Spec.Containers[0].SecurityContext = &v1.SecurityContext{}
		deployment.Spec.Template.Spec.ImagePullSecrets = []v1.LocalObjectReference{{
			// TODO make this as part of operator in a safe way
			Name: "github-auth",
		}}
		deployment.Spec.Template.Spec.SecurityContext = &v1.PodSecurityContext{}
		deployment.Spec.Template.Spec.Volumes = volumes
		deployment.Spec.Template.Spec.Containers[0].VolumeMounts = volumeMounts
	}

	// Re-create list when amount of elements change to flush and apply
	if len(deployment.Spec.Template.Spec.Containers[0].Env) != len(app.Spec.Env) {
		deployment.Spec.Template.Spec.Containers[0].Env = make([]v1.EnvVar, len(app.Spec.Env))
	}

	// Re-create list when amount of elements change to flush and apply
	if len(deployment.Spec.Template.Spec.Containers[0].EnvFrom) != len(app.Spec.EnvFrom) {
		deployment.Spec.Template.Spec.Containers[0].EnvFrom = make([]v1.EnvFromSource, len(app.Spec.EnvFrom))
	}

	deployment.Spec.Replicas = &replicas
	deployment.Spec.Template.ObjectMeta.Labels = labels

	deployment.Spec.Template.Spec.SecurityContext.SupplementalGroups = []int64{uid}
	deployment.Spec.Template.Spec.SecurityContext.FSGroup = &uid

	deployment.Spec.Template.Spec.Containers[0].Name = app.Name
	deployment.Spec.Template.Spec.Containers[0].Image = app.Spec.Image
	deployment.Spec.Template.Spec.Containers[0].ImagePullPolicy = v1.PullAlways

	if deployment.Spec.Template.Spec.Containers[0].SecurityContext.SeccompProfile == nil {
		deployment.Spec.Template.Spec.Containers[0].SecurityContext.SeccompProfile = &v1.SeccompProfile{}
	}
	deployment.Spec.Template.Spec.Containers[0].SecurityContext.SeccompProfile.Type = "RuntimeDefault"
	deployment.Spec.Template.Spec.Containers[0].SecurityContext.Privileged = &no
	deployment.Spec.Template.Spec.Containers[0].SecurityContext.AllowPrivilegeEscalation = &no
	deployment.Spec.Template.Spec.Containers[0].SecurityContext.ReadOnlyRootFilesystem = &yes
	deployment.Spec.Template.Spec.Containers[0].SecurityContext.RunAsUser = &uid
	deployment.Spec.Template.Spec.Containers[0].SecurityContext.RunAsGroup = &uid

	deployment.Spec.Template.Spec.Containers[0].Ports[0].ContainerPort = int32(app.Spec.Port)
	deployment.Spec.Template.Spec.Containers[0].Resources = reconciler.buildResources(ctx, app)

	reconciler.addEnvData(app, deployment.Spec.Template.Spec.Containers[0].Env)
	reconciler.addEnvFromData(app, deployment.Spec.Template.Spec.Containers[0].EnvFrom)
	reconciler.addProbes(app, &deployment.Spec.Template.Spec.Containers[0])
}

func (*ApplicationReconciler) addProbes(app *skiperatorv1alpha1.Application, container *v1.Container) {
	if app.Spec.Liveness != nil {
		if container.LivenessProbe == nil {
			container.LivenessProbe = &v1.Probe{}
			container.LivenessProbe.ProbeHandler = v1.ProbeHandler{
				HTTPGet: &v1.HTTPGetAction{},
			}
		}
		// When unset in config (0) k8s sets default values which are different
		// Prevent infinite update loop by only setting when non-null
		if app.Spec.Liveness.FailureThreshold > 0 {
			container.LivenessProbe.FailureThreshold = int32(app.Spec.Liveness.FailureThreshold)
		}
		if app.Spec.Liveness.InitialDelay > 0 {
			container.LivenessProbe.InitialDelaySeconds = int32(app.Spec.Liveness.InitialDelay)
		}
		if app.Spec.Liveness.Timeout > 0 {
			container.LivenessProbe.TimeoutSeconds = int32(app.Spec.Liveness.Timeout)
		}
		container.LivenessProbe.ProbeHandler.HTTPGet.Path = app.Spec.Liveness.Path
		container.LivenessProbe.ProbeHandler.HTTPGet.Port = intstr.FromInt(app.Spec.Liveness.Port)
	} else {
		container.LivenessProbe = nil
	}

	if app.Spec.Readiness != nil {
		if container.ReadinessProbe == nil {
			container.ReadinessProbe = &v1.Probe{}
			container.ReadinessProbe.ProbeHandler = v1.ProbeHandler{
				HTTPGet: &v1.HTTPGetAction{},
			}
		}
		// When unset in config (0) k8s sets default values which are different
		// Prevent infinite update loop by only setting when non-null
		if app.Spec.Readiness.FailureThreshold > 0 {
			container.ReadinessProbe.FailureThreshold = int32(app.Spec.Readiness.FailureThreshold)
		}
		if app.Spec.Readiness.InitialDelay > 0 {
			container.ReadinessProbe.InitialDelaySeconds = int32(app.Spec.Readiness.InitialDelay)
		}
		if app.Spec.Readiness.Timeout > 0 {
			container.ReadinessProbe.TimeoutSeconds = int32(app.Spec.Readiness.Timeout)
		}
		container.ReadinessProbe.ProbeHandler.HTTPGet.Path = app.Spec.Readiness.Path
		container.ReadinessProbe.ProbeHandler.HTTPGet.Port = intstr.FromInt(app.Spec.Liveness.Port)
	} else {
		container.ReadinessProbe = nil
	}
}

func (*ApplicationReconciler) addEnvData(app *skiperatorv1alpha1.Application, envList []v1.EnvVar) {
	for i, env := range app.Spec.Env {
		envList[i].Name = env.Name
		envList[i].Value = env.Value

		if envList[i].ValueFrom == nil {
			envList[i].ValueFrom = env.ValueFrom
		}

		// Allow unsetting
		// TODO support changing the value somehow without infinite update loop
		if envList[i].ValueFrom != nil && env.ValueFrom == nil {
			envList[i].ValueFrom = nil
		}
	}
}

func (*ApplicationReconciler) addEnvFromData(app *skiperatorv1alpha1.Application, envFromList []v1.EnvFromSource) {
	for i, env := range app.Spec.EnvFrom {
		if len(env.Configmap) > 0 {
			if envFromList[i].ConfigMapRef == nil {
				envFromList[i].ConfigMapRef = &v1.ConfigMapEnvSource{
					LocalObjectReference: v1.LocalObjectReference{},
				}
			}
			envFromList[i].ConfigMapRef.LocalObjectReference.Name = env.Configmap
			envFromList[i].SecretRef = nil // sanity check
		} else if len(env.Secret) > 0 {
			if envFromList[i].SecretRef == nil {
				envFromList[i].SecretRef = &v1.SecretEnvSource{
					LocalObjectReference: v1.LocalObjectReference{},
				}
			}
			envFromList[i].SecretRef.LocalObjectReference.Name = env.Secret
			envFromList[i].ConfigMapRef = nil // sanity check
		}
	}
}

func (*ApplicationReconciler) buildVolumes(app *skiperatorv1alpha1.Application) ([]v1.VolumeMount, []v1.Volume) {
	volumeMounts := []v1.VolumeMount{}
	volumes := []v1.Volume{}

	// Add /tmp volume as we specify a read-only root file system and
	// /tmp is commonly used in many containers out of the box
	volumeMounts = append(volumeMounts, v1.VolumeMount{
		Name:      "tmp",
		MountPath: "/tmp",
	})
	volumes = append(volumes, v1.Volume{
		Name: "tmp",
		VolumeSource: v1.VolumeSource{
			EmptyDir: &v1.EmptyDirVolumeSource{
				Medium: v1.StorageMediumMemory,
			},
		},
	})

	// Add volumes specified in FromFiles
	for _, file := range app.Spec.FilesFrom {
		if len(file.Configmap) > 0 {
			volumeMounts = append(volumeMounts, v1.VolumeMount{
				Name:      file.Configmap,
				MountPath: file.MountPath,
			})
			volumes = append(volumes, v1.Volume{
				Name: file.Configmap,
				VolumeSource: v1.VolumeSource{
					ConfigMap: &v1.ConfigMapVolumeSource{
						LocalObjectReference: v1.LocalObjectReference{
							Name: file.Configmap,
						},
					},
				},
			})
		} else if len(file.PersistentVolumeClaim) > 0 {
			volumeMounts = append(volumeMounts, v1.VolumeMount{
				Name:      file.PersistentVolumeClaim,
				MountPath: file.MountPath,
			})
			volumes = append(volumes, v1.Volume{
				Name: file.PersistentVolumeClaim,
				VolumeSource: v1.VolumeSource{
					PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
						ClaimName: file.PersistentVolumeClaim,
					},
				},
			})
		} else if len(file.Secret) > 0 {
			volumeMounts = append(volumeMounts, v1.VolumeMount{
				Name:      file.Secret,
				MountPath: file.MountPath,
			})
			volumes = append(volumes, v1.Volume{
				Name: file.Secret,
				VolumeSource: v1.VolumeSource{
					Secret: &v1.SecretVolumeSource{
						SecretName: file.Secret,
					},
				},
			})
		}
	}
	return volumeMounts, volumes
}

func (reconciler *ApplicationReconciler) buildResources(ctx context.Context, app *skiperatorv1alpha1.Application) v1.ResourceRequirements {
	log := log.FromContext(ctx)
	limits := v1.ResourceList{}
	requests := v1.ResourceList{}

	if app.Spec.Resources == nil {
		return v1.ResourceRequirements{}
	}

	cpuLimit, err := resource.ParseQuantity(app.Spec.Resources.Limits.Cpu)
	if err == nil {
		limits[v1.ResourceCPU] = cpuLimit
	} else if len(app.Spec.Resources.Limits.Cpu) > 0 {
		log.Error(err, "Failed to parse cpu limit object", "input", cpuLimit)
	}

	memLimit, err := resource.ParseQuantity(app.Spec.Resources.Limits.Memory)
	if err == nil {
		limits[v1.ResourceMemory] = memLimit
	} else if len(app.Spec.Resources.Limits.Memory) > 0 {
		log.Error(err, "Failed to parse mem limit object", "input", memLimit)
	}

	cpuRequest, err := resource.ParseQuantity(app.Spec.Resources.Requests.Cpu)
	if err == nil {
		requests[v1.ResourceCPU] = cpuRequest
	} else if len(app.Spec.Resources.Requests.Cpu) > 0 {
		log.Error(err, "Failed to parse cpu request object", "input", cpuRequest)
	}

	memRequest, err := resource.ParseQuantity(app.Spec.Resources.Requests.Memory)
	if err == nil {
		requests[v1.ResourceMemory] = memRequest
	} else if len(app.Spec.Resources.Requests.Memory) > 0 {
		log.Error(err, "Failed to parse mem request object", "input", memRequest)
	}

	return v1.ResourceRequirements{
		Limits:   limits,
		Requests: requests,
	}
}

func (reconciler *ApplicationReconciler) addSidecarDara(app *skiperatorv1alpha1.Application, sidecar *istioNetworkingv1beta1.Sidecar) {
	if sidecar.Spec.OutboundTrafficPolicy == nil {
		sidecar.Spec.OutboundTrafficPolicy = &istioApiNetworkingv1beta1.OutboundTrafficPolicy{}
	}
	// TODO the value below is omitted when viewed in k8s due to JSON
	// omitonly on the OutboundTrafficPolicy struct. Bug in istio API?
	sidecar.Spec.OutboundTrafficPolicy.Mode = istioApiNetworkingv1beta1.OutboundTrafficPolicy_REGISTRY_ONLY
}

func (reconciler *ApplicationReconciler) addEgressNetworkPolicyData(app *skiperatorv1alpha1.Application, networkPolicy *networkingv1.NetworkPolicy) {
	labels := labelsForApplication(app)
	var egressRules []networkingv1.NetworkPolicyEgressRule = networkPolicy.Spec.Egress
	rulesSize := 1 // Always create DNS rule
	shouldCreateOutboundRules := app.Spec.AccessPolicy != nil && app.Spec.AccessPolicy.Outbound != nil && len(app.Spec.AccessPolicy.Outbound.Rules) > 0
	shouldCreateEgressGatewayRule := app.Spec.AccessPolicy != nil && app.Spec.AccessPolicy.Outbound != nil && len(app.Spec.AccessPolicy.Outbound.External) > 0

	// calculate amount of rules
	if app.Spec.AccessPolicy != nil && app.Spec.AccessPolicy.Outbound != nil {
		rulesSize = rulesSize + len(app.Spec.AccessPolicy.Outbound.Rules)
	}

	// Allow traffic to egress when egress traffic is configured
	if shouldCreateEgressGatewayRule {
		rulesSize = rulesSize + 1
	}

	// Initialize array
	if len(egressRules) != rulesSize {
		egressRules = make([]networkingv1.NetworkPolicyEgressRule, rulesSize)
	}

	// Build rules for pods
	if shouldCreateOutboundRules {
		for i, inboundApp := range app.Spec.AccessPolicy.Outbound.Rules {
			namespace := inboundApp.Namespace
			if len(namespace) == 0 {
				namespace = app.Namespace
			}

			if len(egressRules[i].To) != 1 {
				egressRules[i].To = []networkingv1.NetworkPolicyPeer{{
					PodSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{},
					},
					NamespaceSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{},
					},
				}}
			}

			egressRules[i].To[0].PodSelector.MatchLabels["application"] = inboundApp.Application
			egressRules[i].To[0].NamespaceSelector.MatchLabels["kubernetes.io/metadata.name"] = namespace
		}
	}

	// Build rule for allowing DNS traffic
	i := rulesSize - 1
	if len(egressRules[i].To) != 1 {
		egressRules[i].To = []networkingv1.NetworkPolicyPeer{{
			PodSelector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"k8s-app": "kube-dns",
				},
			},
			NamespaceSelector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"kubernetes.io/metadata.name": "kube-system",
				},
			},
		}}
	}

	if len(egressRules[i].Ports) != 2 {
		egressRules[i].Ports = make([]networkingv1.NetworkPolicyPort, 2)
	}

	dnsPort := intstr.FromInt(53)
	udp := v1.ProtocolUDP
	tcp := v1.ProtocolTCP
	egressRules[i].Ports[0].Protocol = &udp
	egressRules[i].Ports[0].Port = &dnsPort
	egressRules[i].Ports[1].Protocol = &tcp
	egressRules[i].Ports[1].Port = &dnsPort

	// Allow traffic to egress when egress traffic is configured
	if shouldCreateEgressGatewayRule {
		i = rulesSize - 2
		egressRules[i].To = []networkingv1.NetworkPolicyPeer{{
			PodSelector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"egress": "external",
				},
			},
			NamespaceSelector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"kubernetes.io/metadata.name": "istio-system",
				},
			},
		}}
	}

	if len(networkPolicy.Spec.PolicyTypes) != 1 {
		networkPolicy.Spec.PolicyTypes = []networkingv1.PolicyType{networkingv1.PolicyTypeEgress}
	}

	networkPolicy.Spec.PodSelector.MatchLabels = labels
	networkPolicy.Spec.Egress = egressRules
}

func (reconciler *ApplicationReconciler) addIngressNetworkPolicyData(app *skiperatorv1alpha1.Application, networkPolicy *networkingv1.NetworkPolicy) {
	labels := labelsForApplication(app)
	port := intstr.FromInt(app.Spec.Port)
	var ingressRules []networkingv1.NetworkPolicyIngressRule = networkPolicy.Spec.Ingress
	rulesSize := 0

	// Add rule for ingress traffic when exposed with hostname
	if len(app.Spec.Ingresses) > 0 {
		rulesSize = rulesSize + 1
	}

	if app.Spec.AccessPolicy != nil && app.Spec.AccessPolicy.Inbound != nil {
		rulesSize = rulesSize + len(app.Spec.AccessPolicy.Inbound.Rules)
	}

	// Initialize
	if len(ingressRules) != rulesSize {
		ingressRules = make([]networkingv1.NetworkPolicyIngressRule, rulesSize)
	}

	// Build rules for pods
	if app.Spec.AccessPolicy != nil && app.Spec.AccessPolicy.Inbound != nil && app.Spec.AccessPolicy.Inbound.Rules != nil {
		for i, inboundApp := range app.Spec.AccessPolicy.Inbound.Rules {
			namespace := inboundApp.Namespace
			if len(namespace) == 0 {
				namespace = app.Namespace
			}

			if len(ingressRules[i].From) != 2 {
				ingressRules[i].From = []networkingv1.NetworkPolicyPeer{{
					PodSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{},
					},
				}, {
					NamespaceSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{},
					},
				}}
			}

			ingressRules[i].From[0].PodSelector.MatchLabels["application"] = inboundApp.Application
			ingressRules[i].From[1].NamespaceSelector.MatchLabels["kubernetes.io/metadata.name"] = namespace
		}
	}

	// Build rule for ingress gateway
	if len(app.Spec.Ingresses) > 0 {
		i := rulesSize - 1
		if len(ingressRules[i].From) != 2 {
			ingressRules[i].From = []networkingv1.NetworkPolicyPeer{{
				PodSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"ingress": "external",
					},
				},
			}, {
				NamespaceSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"kubernetes.io/metadata.name": "istio-system",
					},
				},
			}}
		}

		if len(ingressRules[i].Ports) != 1 {
			ingressRules[i].Ports = make([]networkingv1.NetworkPolicyPort, 1)
		}
		ingressRules[i].Ports[0].Port = &port
	}

	if len(networkPolicy.Spec.PolicyTypes) != 1 {
		networkPolicy.Spec.PolicyTypes = []networkingv1.PolicyType{networkingv1.PolicyTypeIngress}
	}

	networkPolicy.Spec.PodSelector.MatchLabels = labels
	networkPolicy.Spec.Ingress = ingressRules
}

func (reconciler *ApplicationReconciler) addPeerAuthenticationData(app *skiperatorv1alpha1.Application, peerAuthentication *istioSecurityv1beta1.PeerAuthentication) {
	if peerAuthentication.Spec.Mtls == nil {
		peerAuthentication.Spec.Mtls = &istioApiSecurityv1beta1.PeerAuthentication_MutualTLS{}
	}
	peerAuthentication.Spec.Mtls.Mode = istioApiSecurityv1beta1.PeerAuthentication_MutualTLS_STRICT
}

// returns the labels for selecting the resources
// belonging to the given CRD name.
func labelsForApplication(app *skiperatorv1alpha1.Application) map[string]string {
	return map[string]string{"application": app.Name}
}

// SetupWithManager sets up the controller with the Manager.
func (reconciler *ApplicationReconciler) SetupWithManager(manager ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(manager).
		For(&skiperatorv1alpha1.Application{}).
		Owns(&appsv1.Deployment{}).
		Owns(&autoscalingv1.HorizontalPodAutoscaler{}).
		Owns(&v1.Service{}).
		Owns(&istioNetworkingv1beta1.Gateway{}).
		Owns(&istioNetworkingv1beta1.VirtualService{}).
		Owns(&istioNetworkingv1beta1.ServiceEntry{}).
		Owns(&networkingv1.NetworkPolicy{}).
		Owns(&istioSecurityv1beta1.PeerAuthentication{}).
		Complete(reconciler)
}
