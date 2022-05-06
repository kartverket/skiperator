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
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
)

// ApplicationReconciler reconciles a Application object
type ApplicationReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=skiperator.kartverket.no,resources=applications,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=skiperator.kartverket.no,resources=applications/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=skiperator.kartverket.no,resources=applications/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Application object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
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

	log.Info("The incoming application object is", "Application", app)

	// Check if the deployment already exists, if not create a new one
	existingDeployment := &appsv1.Deployment{}
	newDeployment := reconciler.buildDeployment(app)
	shouldReturn, result, err := reconciler.installObject(ctx, app, existingDeployment, newDeployment)
	if shouldReturn {
		return result, err
	}

	// Check if the deployment already exists, if not create a new one
	existingService := &v1.Service{}
	newService := reconciler.buildService(app)
	shouldReturn, result, err = reconciler.installObject(ctx, app, existingService, newService)
	if shouldReturn {
		return result, err
	}

	// TODO make service

	// TODO make Gateway

	// TODO make VirtualService

	// TODO make ResourceLimit

	// TODO make autoscaling

	// TODO make image pull Secret

	// Check if the networkPolicy already exists, if not create a new one
	existingNetworkPolicy := &networkingv1.NetworkPolicy{}
	newNetworkPolicy := reconciler.buildNetworkPolicy(app)
	shouldReturn, result, err = reconciler.installObject(ctx, app, existingNetworkPolicy, newNetworkPolicy)
	if shouldReturn {
		return result, err
	}

	// Check if the peerAuthentication already exists, if not create a new one
	existingPeerAuthentication := &istioSecurityv1beta1.PeerAuthentication{}
	newPeerAuthencitaion := reconciler.buildPeerAuthentication(app)
	shouldReturn, result, err = reconciler.installObject(ctx, app, existingPeerAuthentication, newPeerAuthencitaion)
	if shouldReturn {
		return result, err
	}

	// Check if the Sidecar already exists, if not create a new one
	existingSidecar := &istioNetworkingv1beta1.Sidecar{}
	newSidecar := reconciler.buildSidecar(app)
	shouldReturn, result, err = reconciler.installObject(ctx, app, existingSidecar, newSidecar)
	if shouldReturn {
		return result, err
	}

	return ctrl.Result{}, err
}

func (reconciler *ApplicationReconciler) installObject(ctx context.Context, app *skiperatorv1alpha1.Application, existingObject client.Object, newObject client.Object) (bool, reconcile.Result, error) {
	log := log.FromContext(ctx)
	err := reconciler.Get(ctx, types.NamespacedName{Name: app.Name, Namespace: app.Namespace}, existingObject)

	if err != nil && errors.IsNotFound(err) {
		// TODO: Get Kind from object here
		kind := "Object"
		namespace := newObject.GetNamespace()
		name := newObject.GetName()

		log.Info("Creating a new "+kind, "newObject.Namespace", namespace, "newObject.Name", name)
		// TODO Look into using ctrl.CreateOrUpdate to make code less imperative
		err = reconciler.Create(ctx, newObject)

		if err != nil {
			log.Error(err, "Failed to create new "+kind, "newObject.Namespace", newObject.GetNamespace(), "newObject.Name", newObject.GetName())
			return true, ctrl.Result{}, err
		}

		return true, ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get existing object")
		return true, ctrl.Result{}, err
	}

	return false, reconcile.Result{}, nil
}

func (reconciler *ApplicationReconciler) buildService(app *skiperatorv1alpha1.Application) *v1.Service {
	labels := labelsForApplication(app)

	service := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      app.Name,
			Namespace: app.Namespace,
		},
		Spec: v1.ServiceSpec{
			Selector: labels,
			Type:     v1.ServiceTypeClusterIP,
			Ports: []v1.ServicePort{{
				Port:       int32(app.Spec.Port),
				TargetPort: intstr.FromInt(app.Spec.Port),
			}},
		},
	}

	// Setting controller as owner makes the NetworkPolicy garbage collected when Application gets deleted in k8s
	ctrl.SetControllerReference(app, service, reconciler.Scheme)
	return service
}

func (reconciler *ApplicationReconciler) buildDeployment(app *skiperatorv1alpha1.Application) *appsv1.Deployment {
	labels := labelsForApplication(app)
	var uid int64 = 150
	yes := true
	no := false

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      app.Name,
			Namespace: app.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: app.Spec.Replicas.Min,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
					Annotations: map[string]string{
						"prometheus.io/scrape":                     "true",
						"seccomp.security.alpha.kubernetes.io/pod": "runtime/default",
					},
				},
				Spec: v1.PodSpec{
					SecurityContext: &v1.PodSecurityContext{
						SupplementalGroups: []int64{uid},
						FSGroup:            &uid,
					},
					ImagePullSecrets: []v1.LocalObjectReference{{
						Name: "github-auth",
					}},
					Containers: []v1.Container{{
						Name:            app.Name,
						Image:           app.Spec.Image,
						ImagePullPolicy: v1.PullAlways,
						SecurityContext: &v1.SecurityContext{
							Privileged:               &no,
							AllowPrivilegeEscalation: &no,
							ReadOnlyRootFilesystem:   &yes,
							RunAsUser:                &uid,
							RunAsGroup:               &uid,
						},
						Ports: []v1.ContainerPort{{
							Name:          "main",
							ContainerPort: int32(app.Spec.Port),
						}},
						// TODO add env
						// TODO add envFrom
					}},
				},
			},
		},
	}

	// Setting controller as owner makes the NetworkPolicy garbage collected when Application gets deleted in k8s
	ctrl.SetControllerReference(app, deployment, reconciler.Scheme)
	return deployment
}

func (reconciler *ApplicationReconciler) buildSidecar(app *skiperatorv1alpha1.Application) *istioNetworkingv1beta1.Sidecar {
	sidecar := &istioNetworkingv1beta1.Sidecar{
		ObjectMeta: metav1.ObjectMeta{
			Name:      app.Name,
			Namespace: app.Namespace,
		},
		Spec: istioApiNetworkingv1beta1.Sidecar{
			OutboundTrafficPolicy: &istioApiNetworkingv1beta1.OutboundTrafficPolicy{
				// TODO the value below is omitted when viewed in k8s due to JSON
				// omitonly on the OutboundTrafficPolicy struct. Bug in istio API?
				Mode: istioApiNetworkingv1beta1.OutboundTrafficPolicy_REGISTRY_ONLY,
			},
		},
	}

	// Setting controller as owner makes the NetworkPolicy garbage collected when Application gets deleted in k8s
	ctrl.SetControllerReference(app, sidecar, reconciler.Scheme)
	return sidecar
}

func (reconciler *ApplicationReconciler) buildNetworkPolicy(app *skiperatorv1alpha1.Application) *networkingv1.NetworkPolicy {
	labels := labelsForApplication(app)
	ingressRules := buildIngressPolicy(app)

	for _, inboundApp := range app.Spec.AccessPolicy.Inbound.Rules {
		ingressRules = append(ingressRules, buildIngressRules(app, inboundApp)...)
	}

	policy := &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      app.Name,
			Namespace: app.Namespace,
		},
		Spec: networkingv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{
				MatchLabels: labels,
			},
			Ingress: ingressRules,
		},
	}

	// Setting controller as owner makes the NetworkPolicy garbage collected when Application gets deleted in k8s
	ctrl.SetControllerReference(app, policy, reconciler.Scheme)
	return policy
}

func buildIngressPolicy(app *skiperatorv1alpha1.Application) []networkingv1.NetworkPolicyIngressRule {
	rule := []networkingv1.NetworkPolicyIngressRule{}

	// When ingresses are set, allow traffic from ingressgateway
	if len(app.Spec.Ingresses) > 0 {
		port := intstr.FromInt(app.Spec.Port)
		rule = append(rule, networkingv1.NetworkPolicyIngressRule{
			From: []networkingv1.NetworkPolicyPeer{{
				NamespaceSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"kubernetes.io/metadata.name": "istio-system",
					},
				},
				PodSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"ingress": "external",
					},
				},
			}},
			Ports: []networkingv1.NetworkPolicyPort{{
				Port: &port,
			}},
		})
	}

	return rule
}

func buildIngressRules(app *skiperatorv1alpha1.Application, inboundApp skiperatorv1alpha1.Rule) []networkingv1.NetworkPolicyIngressRule {
	rule := []networkingv1.NetworkPolicyIngressRule{}

	// Add ingress rule for app
	namespace := inboundApp.Namespace
	if len(namespace) == 0 {
		namespace = app.Namespace
	}
	rule = append(rule, networkingv1.NetworkPolicyIngressRule{
		From: []networkingv1.NetworkPolicyPeer{{
			PodSelector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"application": inboundApp.Application,
				},
			},
			NamespaceSelector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"kubernetes.io/metadata.name": namespace,
				},
			},
		}},
	})

	return rule
}

func (reconciler *ApplicationReconciler) buildPeerAuthentication(app *skiperatorv1alpha1.Application) *istioSecurityv1beta1.PeerAuthentication {
	peerAuthentication := istioSecurityv1beta1.PeerAuthentication{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: app.Namespace,
			Name:      app.Name,
		},
		Spec: istioApiSecurityv1beta1.PeerAuthentication{
			Mtls: &istioApiSecurityv1beta1.PeerAuthentication_MutualTLS{
				Mode: istioApiSecurityv1beta1.PeerAuthentication_MutualTLS_STRICT,
			},
		},
	}

	// Setting controller as owner makes the PeerAuthentication garbage collected when Application gets deleted in k8s
	ctrl.SetControllerReference(app, &peerAuthentication, reconciler.Scheme)
	return &peerAuthentication
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
		Owns(&networkingv1.NetworkPolicy{}).
		Owns(&istioSecurityv1beta1.PeerAuthentication{}).
		Complete(reconciler)
}
