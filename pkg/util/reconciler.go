package util

import (
	"context"
	"fmt"
	"github.com/kartverket/skiperator/pkg/flags"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/kartverket/skiperator/api/v1alpha1/podtypes"
	corev1 "k8s.io/api/core/v1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// ReconcilerBase is a base struct from which all reconcilers can be derived from. By doing so your reconcilers will also inherit a set of utility functions
// To inherit from reconciler just build your finalizer this way:
//
//	type MyReconciler struct {
//	  util.ReconcilerBase
//	  ... other optional fields ...
//	}
type ReconcilerBase struct {
	apireader        client.Reader
	client           client.Client
	extensionsClient *apiextensionsclient.Clientset
	scheme           *runtime.Scheme
	restConfig       *rest.Config
	recorder         record.EventRecorder
	features         *flags.Features
}

type Controller interface {
	SetupWithManager() error
}

func NewReconcilerBase(client client.Client, extensionsClient *apiextensionsclient.Clientset, scheme *runtime.Scheme, restConfig *rest.Config, recorder record.EventRecorder, apireader client.Reader) ReconcilerBase {
	return ReconcilerBase{
		apireader:        apireader,
		client:           client,
		extensionsClient: extensionsClient,
		scheme:           scheme,
		restConfig:       restConfig,
		recorder:         recorder,
		features:         flags.FeatureFlags,
	}
}

// NewReconcilerBase is a construction function to create a new ReconcilerBase.
func NewFromManager(mgr manager.Manager, recorder record.EventRecorder) ReconcilerBase {
	extensionsClient, err := apiextensionsclient.NewForConfig(mgr.GetConfig())
	if err != nil {
		ctrl.Log.Error(err, "could not create extensions client, won't be able to peek at CRDs")
	}

	return NewReconcilerBase(mgr.GetClient(), extensionsClient, mgr.GetScheme(), mgr.GetConfig(), recorder, mgr.GetAPIReader())
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

func (r *ReconcilerBase) GetEgressServices(ctx context.Context, owner client.Object, accessPolicy *podtypes.AccessPolicy) ([]corev1.Service, error) {
	var egressServices []corev1.Service
	if accessPolicy == nil {
		return egressServices, nil
	}

	for _, outboundRule := range accessPolicy.Outbound.Rules {
		namespaces := []string{}

		if outboundRule.Namespace != "" {
			namespaces = []string{outboundRule.Namespace}
		} else if outboundRule.NamespacesByLabel == nil {
			if outboundRule.Namespace == "" {
				namespaces = []string{owner.GetNamespace()}
			}

		} else {
			namespaceList := corev1.NamespaceList{}

			err := r.GetClient().List(ctx, &namespaceList, client.MatchingLabels(outboundRule.NamespacesByLabel))

			if errors.IsNotFound(err) {
				r.EmitWarningEvent(owner, "NoNamespaces", fmt.Sprintf("cannot find any namespaces"))
				return egressServices, err
			} else if err != nil {
				return egressServices, err
			}

			for _, namespace := range namespaceList.Items {
				namespaces = append(namespaces, namespace.Name)
			}

		}

		for _, namespace := range namespaces {
			service := corev1.Service{}

			err := r.GetClient().Get(ctx, client.ObjectKey{
				Namespace: namespace,
				Name:      outboundRule.Application,
			}, &service)
			if errors.IsNotFound(err) {
				r.EmitWarningEvent(owner, "MissingApplication", fmt.Sprintf("cannot find Application named %s in Namespace %s, egress rule will not be added", outboundRule.Application, outboundRule.Namespace))
				continue
			} else if err != nil {
				return egressServices, err
			}

			egressServices = append(egressServices, service)
		}
	}

	return egressServices, nil
}

func (r *ReconcilerBase) GetNamespaces(ctx context.Context, owner client.Object) (corev1.NamespaceList, error) {
	namespaces := corev1.NamespaceList{}

	err := r.GetClient().List(ctx, &namespaces)

	if errors.IsNotFound(err) {
		r.EmitWarningEvent(owner, "NoNamespaces", fmt.Sprintf("cannot find any namespaces"))
		return namespaces, err
	} else if err != nil {
		return namespaces, err
	}

	return namespaces, nil
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

	v, exists := namespace.Labels[IstioRevisionLabel]

	return exists && len(v) > 0
}

func hasIgnoreLabel(obj client.Object) bool {
	labels := obj.GetLabels()
	return labels["skiperator.kartverket.no/ignore"] == "true"
}

func (r *ReconcilerBase) ShouldReconcile(ctx context.Context, obj client.Object) (bool, error) {
	copyObj := obj.DeepCopyObject().(client.Object)
	err := r.GetClient().Get(ctx, client.ObjectKeyFromObject(copyObj), copyObj)
	err = client.IgnoreNotFound(err)

	if err != nil {
		return false, err
	}

	shouldReconcile := !hasIgnoreLabel(copyObj)

	return shouldReconcile, nil
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

func (r *ReconcilerBase) DeleteObjectIfExists(ctx context.Context, object client.Object) error {
	err := client.IgnoreNotFound(r.GetClient().Delete(ctx, object))
	if err != nil {
		return err
	}

	return nil
}

func doNotRequeue() (reconcile.Result, error) {
	return reconcile.Result{}, nil
}

func RequeueWithError(err error) (reconcile.Result, error) {
	return reconcile.Result{}, err
}
