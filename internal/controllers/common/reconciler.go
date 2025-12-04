package common

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/kartverket/skiperator/api/common/podtypes"
	"github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/resourceutils"
	"github.com/kartverket/skiperator/pkg/util"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// ReconcilerBase is a base struct from which all reconcilers can be derived from. By doing so your reconcilers will also inherit a set of utility functions
// To inherit the functionality just build your reconciler this way:
//
//	type MyReconciler struct {
//	  util.ReconcilerBase
//	  ... other optional fields ...
//	}
type ReconcilerBase struct {
	client           client.Client
	extensionsClient *apiextensionsclient.Clientset
	scheme           *runtime.Scheme
	restConfig       *rest.Config
	recorder         record.EventRecorder
	Logger           logr.Logger
}

func NewReconcilerBase(
	client client.Client,
	extensionsClient *apiextensionsclient.Clientset,
	scheme *runtime.Scheme,
	restConfig *rest.Config,
	recorder record.EventRecorder,
) ReconcilerBase {
	return ReconcilerBase{
		client:           client,
		extensionsClient: extensionsClient,
		scheme:           scheme,
		restConfig:       restConfig,
		recorder:         recorder,
	}
}

func NewFromManager(mgr manager.Manager, recorder record.EventRecorder) ReconcilerBase {
	extensionsClient, err := apiextensionsclient.NewForConfig(mgr.GetConfig())
	if err != nil {
		ctrl.Log.Error(err, "could not create extensions client, won't be able to peek at CRDs")
	}

	return NewReconcilerBase(mgr.GetClient(), extensionsClient, mgr.GetScheme(), mgr.GetConfig(), recorder)
}

// GetClient returns the underlying client
func (r *ReconcilerBase) GetClient() client.Client {
	return r.client
}

// GetApiExtensionsClient returns the underlying API Extensions client
func (r *ReconcilerBase) GetApiExtensionsClient() *apiextensionsclient.Clientset {
	return r.extensionsClient
}

// GetRestConfig returns the underlying rest config
func (r *ReconcilerBase) GetRestConfig() *rest.Config {
	return r.restConfig
}

// GetRecorder returns the underlying recorder
func (r *ReconcilerBase) GetRecorder() record.EventRecorder {
	return r.recorder
}

// GetScheme returns the scheme
func (r *ReconcilerBase) GetScheme() *runtime.Scheme {
	return r.scheme
}

func (r *ReconcilerBase) EmitWarningEvent(object runtime.Object, reason string, message string) {
	r.GetRecorder().Event(
		object,
		corev1.EventTypeWarning, reason,
		message,
	)
}

func (r *ReconcilerBase) EmitNormalEvent(object runtime.Object, reason string, message string) {
	r.GetRecorder().Event(
		object,
		corev1.EventTypeNormal, reason,
		message,
	)
}

func (r *ReconcilerBase) IsIstioEnabledForNamespace(ctx context.Context, namespaceName string) bool {
	namespace := corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespaceName,
		},
	}

	err := r.GetClient().Get(ctx, client.ObjectKeyFromObject(&namespace), &namespace)
	if err != nil {
		return false
	}

	v, exists := namespace.Labels[util.IstioRevisionLabel]

	return exists && len(v) > 0
}

func (r *ReconcilerBase) SetSubresourceDefaults(resources []client.Object, skipObj client.Object) error {
	for _, resource := range resources {
		if err := resourceutils.AddGVK(r.GetScheme(), resource); err != nil {
			return err
		}
		resourceutils.SetCommonAnnotations(resource)
		if err := resourceutils.SetOwnerReference(skipObj, resource, r.GetScheme()); err != nil {
			return err
		}
	}
	return nil
}

func (r *ReconcilerBase) SetErrorState(ctx context.Context, skipObj v1alpha1.SKIPObject, err error, message string, reason string) {
	r.EmitWarningEvent(skipObj, reason, message)
	skipObj.GetStatus().SetSummaryError(message + ": " + err.Error())
	r.updateStatus(ctx, skipObj)
}

func (r *ReconcilerBase) SetProgressingState(ctx context.Context, skipObj v1alpha1.SKIPObject, message string) {
	r.EmitNormalEvent(skipObj, "ReconcileStart", message)
	skipObj.GetStatus().SetSummaryProgressing()
	r.updateStatus(ctx, skipObj)
}

func (r *ReconcilerBase) SetSyncedState(ctx context.Context, skipObj v1alpha1.SKIPObject, message string) {
	r.EmitNormalEvent(skipObj, "ReconcileEndSuccess", message)
	skipObj.GetStatus().SetSummarySynced()
	r.updateStatus(ctx, skipObj)
}

func (r *ReconcilerBase) updateStatus(ctx context.Context, skipObj v1alpha1.SKIPObject) {
	latestObj := skipObj.DeepCopyObject().(v1alpha1.SKIPObject)
	key := client.ObjectKeyFromObject(skipObj)

	if err := r.GetClient().Get(ctx, key, latestObj); err != nil {
		r.Logger.Error(err, "Failed to get latest object version")
	}
	latestObj.SetStatus(*skipObj.GetStatus())
	if err := r.GetClient().Status().Update(ctx, latestObj); err != nil {
		r.Logger.Error(err, "Failed to update status")
	}
}

func (r *ReconcilerBase) getTargetApplicationPorts(ctx context.Context, appName string, namespace string) ([]networkingv1.NetworkPolicyPort, error) {
	service := &corev1.Service{}
	var servicePorts []networkingv1.NetworkPolicyPort

	if err := r.GetClient().Get(ctx, types.NamespacedName{Name: appName, Namespace: namespace}, service); err != nil {
		if errors.IsNotFound(err) {
			return servicePorts, nil
		}
		return nil, fmt.Errorf("error when trying to get target application: %s", err.Error())
	}

	for _, port := range service.Spec.Ports {
		servicePorts = append(servicePorts, networkingv1.NetworkPolicyPort{
			Port: util.PointTo(intstr.FromInt32(port.Port)),
		})
	}
	return servicePorts, nil
}

func (r *ReconcilerBase) UpdateAccessPolicy(ctx context.Context, obj v1alpha1.SKIPObject) {
	if obj.GetCommonSpec().AccessPolicy == nil {
		return
	}

	if obj.GetCommonSpec().AccessPolicy.Outbound != nil {
		if errs := r.setPortsForRules(ctx, obj.GetCommonSpec().AccessPolicy.Outbound.Rules, obj.GetNamespace()); len(errs) != 0 {
			for _, err := range errs {
				r.EmitWarningEvent(obj, "InvalidAccessPolicy", fmt.Sprintf("failed to set ports for outbound rules: %s", err.Error()))
			}
		}
	}
}

func (r *ReconcilerBase) setPortsForRules(ctx context.Context, rules []podtypes.InternalRule, skipObjNamespace string) []error {
	var ruleErrors []error
	for i := range rules {
		rule := &rules[i]
		if len(rule.Ports) != 0 {
			continue
		}
		var namespaceList []string
		switch {
		case rule.Namespace != "":
			namespaceList = append(namespaceList, rule.Namespace)
		case len(rule.NamespacesByLabel) != 0:
			namespaces, err := r.GetNamespacesByLabel(ctx, rule)
			if err != nil {
				ruleErrors = append(ruleErrors, err)
			}
			for _, ns := range namespaces.Items {
				namespaceList = append(namespaceList, ns.Name)
			}
		default:
			namespaceList = append(namespaceList, skipObjNamespace)
		}

		if len(namespaceList) == 0 {
			ruleErrors = append(ruleErrors, fmt.Errorf("expected namespace, but found none for application %s", rule.Application))
		}

		for _, ns := range namespaceList {
			targetAppPorts, err := r.getTargetApplicationPorts(ctx, rule.Application, ns)
			if err != nil {
				ruleErrors = append(ruleErrors, err)
			}
			if len(targetAppPorts) == 0 {
				ruleErrors = append(ruleErrors, fmt.Errorf("no ports found for application %s in namespace %s", rule.Application, ns))
				continue
			}
			rule.Ports = append(rule.Ports, targetAppPorts...)
		}
	}
	return ruleErrors
}

func (r *ReconcilerBase) GetNamespacesByLabel(ctx context.Context, rule *podtypes.InternalRule) (*corev1.NamespaceList, error) {
	namespaces := &corev1.NamespaceList{}
	selector := metav1.LabelSelector{MatchLabels: rule.NamespacesByLabel}
	selectorString, err := metav1.LabelSelectorAsSelector(&selector)
	if err != nil {
		return namespaces, fmt.Errorf("failed to create label selector: %w", err)
	}
	if err = r.GetClient().List(ctx, namespaces, &client.ListOptions{LabelSelector: selectorString}); err != nil {
		return namespaces, fmt.Errorf("failed to list namespaces: %w", err)
	}
	return namespaces, nil
}
