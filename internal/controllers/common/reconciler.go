package common

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/kartverket/skiperator/pkg/resourceprocessor"
	"github.com/kartverket/skiperator/pkg/resourceschemas"
	"github.com/kartverket/skiperator/pkg/util"
	corev1 "k8s.io/api/core/v1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
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
	client           client.Client
	extensionsClient *apiextensionsclient.Clientset
	scheme           *runtime.Scheme
	restConfig       *rest.Config
	recorder         record.EventRecorder
	processor        *resourceprocessor.ResourceProcessor
	Logger           logr.Logger
}

func NewReconcilerBase(
	client client.Client,
	extensionsClient *apiextensionsclient.Clientset,
	scheme *runtime.Scheme,
	restConfig *rest.Config,
	recorder record.EventRecorder,
	processor *resourceprocessor.ResourceProcessor,
) ReconcilerBase {
	return ReconcilerBase{
		client:           client,
		extensionsClient: extensionsClient,
		scheme:           scheme,
		restConfig:       restConfig,
		recorder:         recorder,
		processor:        processor,
	}
}

// NewReconcilerBase is a construction function to create a new ReconcilerBase.
func NewFromManager(mgr manager.Manager, recorder record.EventRecorder) ReconcilerBase {
	extensionsClient, err := apiextensionsclient.NewForConfig(mgr.GetConfig())
	if err != nil {
		ctrl.Log.Error(err, "could not create extensions client, won't be able to peek at CRDs")
	}
	//TODO needs to be schemas for each type
	processor := resourceprocessor.NewResourceProcessor(mgr.GetClient(), resourceschemas.GetApplicationSchemas(mgr.GetScheme()), mgr.GetScheme())

	return NewReconcilerBase(mgr.GetClient(), extensionsClient, mgr.GetScheme(), mgr.GetConfig(), recorder, processor)
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

func (r *ReconcilerBase) GetProcessor() *resourceprocessor.ResourceProcessor {
	return r.processor
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

func (r *ReconcilerBase) GetIdentityConfigMap(ctx context.Context) (*corev1.ConfigMap, error) {
	namespacedName := types.NamespacedName{Name: "gcp-identity-config", Namespace: "skiperator-system"}
	identityConfigMap := &corev1.ConfigMap{}
	if err := r.client.Get(ctx, namespacedName, identityConfigMap); err != nil {
		return nil, err
	}
	return identityConfigMap, nil
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
