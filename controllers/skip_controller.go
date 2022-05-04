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

	"istio.io/api/security/v1beta1"
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
)

// SkipReconciler reconciles a Skip object
type SkipReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=skiperator.kartverket.no,resources=skips,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=skiperator.kartverket.no,resources=skips/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=skiperator.kartverket.no,resources=skips/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Skip object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *SkipReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// Lookup the Skip instance for this reconcile request
	skip := &skiperatorv1alpha1.Skip{}
	log.Info("The incoming skip object is", "SKIP", skip)
	err := r.Get(ctx, req.NamespacedName, skip)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info("Skip resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get Skip")
		return ctrl.Result{}, err
	}

	// Check if the networkPolicy already exists, if not create a new one
	existingNetworkPolicy := &networkingv1.NetworkPolicy{}
	err = r.Get(ctx, types.NamespacedName{Name: skip.Name, Namespace: skip.Namespace}, existingNetworkPolicy)
	if err != nil && errors.IsNotFound(err) {
		// Define a new networkPolicy
		networkPolicy := r.buildNetworkPolicy(skip)
		log.Info("Creating a new NetworkPolicy", "NetworkPolicy.Namespace", networkPolicy.Namespace, "NetworkPolicy.Name", networkPolicy.Name)
		err = r.Create(ctx, networkPolicy)
		if err != nil {
			log.Error(err, "Failed to create new NetworkPolicy", "NetworkPolicy.Namespace", networkPolicy.Namespace, "NetworkPolicy.Name", networkPolicy.Name)
			return ctrl.Result{}, err
		}
		// Deployment created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get NetworkPolicy")
		return ctrl.Result{}, err
	}

	// Check if the peerAuthentication already exists, if not create a new one
	existingPeerAuthentication := &securityv1beta1.PeerAuthentication{}
	err = r.Get(ctx, types.NamespacedName{Name: skip.Name, Namespace: skip.Namespace}, existingPeerAuthentication)
	if err != nil && errors.IsNotFound(err) {
		peerAuthencitaion := r.buildPeerAuthentication(skip)
		log.Info("Creating a new PeerAuthentication", "PeerAuthentication.Namespace", peerAuthencitaion.Namespace, "PeerAuthentication.Name", peerAuthencitaion.Name)
		err = r.Create(ctx, peerAuthencitaion)
		if err != nil {
			log.Error(err, "Failed to create new PeerAuthentication", "PeerAuthentication.Namespace", peerAuthencitaion.Namespace, "PeerAuthentication.Name", peerAuthencitaion.Name)
			return ctrl.Result{}, err
		}
		// peerAuthentication created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get PeerAuthentication")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, err
}

func (reconciler *SkipReconciler) buildNetworkPolicy(skip *skiperatorv1alpha1.Skip) *networkingv1.NetworkPolicy {
	ls := labelsForSkip(skip.Name)

	dep := &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      skip.Name,
			Namespace: skip.Namespace,
		},
		Spec: networkingv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{
				MatchLabels: ls,
			},
			Ingress: buildIngressRules(skip),
		},
	}

	// Set instance as the owner and controller
	// ctrl.SetControllerReference(reconciler, dep, reconciler.Scheme)
	return dep
}

// returns the labels for selecting the resources
// belonging to the given CRD name.
func labelsForSkip(name string) map[string]string {
	return map[string]string{"app": "memcached", "memcached_cr": name}
}

func buildIngressRules(skip *skiperatorv1alpha1.Skip) []networkingv1.NetworkPolicyIngressRule {
	rule := []networkingv1.NetworkPolicyIngressRule{}

	for _, policy := range skip.Spec.NetworkPolicies {
		if policy.AcceptIngressTraffic {
			port := intstr.FromInt(8080)
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
	}

	return rule
	/*
		[]networkingv1.NetworkPolicyIngressRule{
			networkingv1.NetworkPolicyIngressRule{
				From: []networkingv1.NetworkPolicyPeer{
					networkingv1.NetworkPolicyPeer{
						NamespaceSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"": "",
							},
						},
					},
					// TODO add iteration for all ingress apps
					networkingv1.NetworkPolicyPeer{
						PodSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								app: "other",
							},
						},
					}
				},
				Ports: []networkingv1.NetworkPolicyPort{
					networkingv1.NetworkPolicyPort{
						Port: 8080,
					}
				}
			},
		},
	*/
}

func (reconciler *SkipReconciler) buildPeerAuthentication(skip *skiperatorv1alpha1.Skip) *securityv1beta1.PeerAuthentication {
	return &securityv1beta1.PeerAuthentication{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: skip.Namespace,
			Name:      skip.Name,
		},
		Spec: v1beta1.PeerAuthentication{
			Mtls: &v1beta1.PeerAuthentication_MutualTLS{
				Mode: v1beta1.PeerAuthentication_MutualTLS_STRICT,
			},
		},
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *SkipReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&skiperatorv1alpha1.Skip{}).
		Owns(&networkingv1.NetworkPolicy{}).
		Owns(&securityv1beta1.PeerAuthentication{}).
		Complete(r)
}
