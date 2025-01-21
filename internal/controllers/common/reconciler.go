package common

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/api/v1alpha1/podtypes"
	"github.com/kartverket/skiperator/pkg/reconciliation"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/resourceutils"
	"github.com/kartverket/skiperator/pkg/resourceprocessor"
	"github.com/kartverket/skiperator/pkg/util"
	"github.com/nais/digdirator/pkg/secrets"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
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

func NewFromManager(mgr manager.Manager, recorder record.EventRecorder, schemas []unstructured.UnstructuredList) ReconcilerBase {
	extensionsClient, err := apiextensionsclient.NewForConfig(mgr.GetConfig())
	if err != nil {
		ctrl.Log.Error(err, "could not create extensions client, won't be able to peek at CRDs")
	}
	processor := resourceprocessor.NewResourceProcessor(mgr.GetClient(), schemas, mgr.GetScheme())

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

func (r *ReconcilerBase) GetAuthConfigsForApplication(ctx context.Context, application *v1alpha1.Application) (*[]reconciliation.AuthConfig, error) {
	identityProviderInfo, err := getIdentityProviderInfoWithAuthenticationEnabled(ctx, application, r.GetClient())
	if err != nil {
		return nil, fmt.Errorf("failed when getting identity provider info: %w", err)
	}
	var authConfigs []reconciliation.AuthConfig
	for _, providerInfo := range *identityProviderInfo {
		switch providerInfo.Provider {
		case reconciliation.ID_PORTEN:
			issuerURI, err := util.GetSecretData(r.GetClient(), ctx, types.NamespacedName{
				Namespace: application.Namespace,
				Name:      providerInfo.SecretName,
			}, secrets.IDPortenIssuerKey)
			if err != nil {
				return nil, fmt.Errorf("failed when retrieving %s: %w", secrets.IDPortenIssuerKey, err)
			}
			jwksURI, err := util.GetSecretData(r.GetClient(), ctx, types.NamespacedName{
				Namespace: application.Namespace,
				Name:      providerInfo.SecretName,
			}, secrets.IDPortenJwksUriKey)
			if err != nil {
				return nil, fmt.Errorf("failed when retrieving %s: %w", secrets.IDPortenJwksUriKey, err)
			}
			clientID, err := util.GetSecretData(r.GetClient(), ctx, types.NamespacedName{
				Namespace: application.Namespace,
				Name:      providerInfo.SecretName,
			}, secrets.IDPortenClientIDKey)
			if err != nil {
				return nil, fmt.Errorf("failed when retrieving %s: %w", secrets.IDPortenClientIDKey, err)
			}
			authConfigs = append(authConfigs, reconciliation.AuthConfig{
				NotPaths: providerInfo.NotPaths,
				ProviderURIs: reconciliation.ProviderURIs{
					Provider:  reconciliation.ID_PORTEN,
					IssuerURI: string(issuerURI),
					JwksURI:   string(jwksURI),
					ClientID:  string(clientID),
				},
			})
		case reconciliation.MASKINPORTEN:
			issuerURI, err := util.GetSecretData(r.GetClient(), ctx, types.NamespacedName{
				Namespace: application.Namespace,
				Name:      providerInfo.SecretName,
			}, secrets.MaskinportenIssuerKey)
			if err != nil {
				return nil, fmt.Errorf("failed when retrieving %s: %w", secrets.MaskinportenIssuerKey, err)
			}
			jwksURI, err := util.GetSecretData(r.GetClient(), ctx, types.NamespacedName{
				Namespace: application.Namespace,
				Name:      providerInfo.SecretName,
			}, secrets.MaskinportenJwksUriKey)
			if err != nil {
				return nil, fmt.Errorf("failed when retrieving %s: %w", secrets.MaskinportenJwksUriKey, err)
			}
			clientID, err := util.GetSecretData(r.GetClient(), ctx, types.NamespacedName{
				Namespace: application.Namespace,
				Name:      providerInfo.SecretName,
			}, secrets.MaskinportenClientIDKey)
			if err != nil {
				return nil, fmt.Errorf("failed when retrieving %s: %w", secrets.MaskinportenClientIDKey, err)
			}
			authConfigs = append(authConfigs, reconciliation.AuthConfig{
				NotPaths: providerInfo.NotPaths,
				ProviderURIs: reconciliation.ProviderURIs{
					Provider:  reconciliation.MASKINPORTEN,
					IssuerURI: string(issuerURI),
					JwksURI:   string(jwksURI),
					ClientID:  string(clientID),
				},
			})
		default:
			return nil, fmt.Errorf("unknown provider: %s", providerInfo.Provider)
		}
	}
	return &authConfigs, nil
}

func getIdentityProviderInfoWithAuthenticationEnabled(ctx context.Context, application *v1alpha1.Application, k8sClient client.Client) (*[]reconciliation.IdentityProviderInfo, error) {
	var providerInfo []reconciliation.IdentityProviderInfo
	if application.Spec.IDPorten != nil && application.Spec.IDPorten.Authentication != nil && application.Spec.IDPorten.Authentication.Enabled {
		var secretName *string
		var err error
		if application.Spec.IDPorten.Authentication.SecretName != nil {
			// If secret name is provided, use it regardless of whether IDPorten is enabled
			secretName = application.Spec.IDPorten.Authentication.SecretName
		} else if application.Spec.IDPorten.Enabled {
			// If IDPorten is enabled but no secretName provided, retrieve the generated secret from IDPortenClient
			secretName, err = getSecretNameForIdentityProvider(k8sClient, ctx,
				types.NamespacedName{
					Namespace: application.Namespace,
					Name:      application.Name,
				},
				reconciliation.ID_PORTEN,
				application.UID)
		} else {
			// If IDPorten is not enabled and no secretName provided, return error
			return nil, fmt.Errorf("JWT authentication requires either IDPorten to be enabled or a secretName to be provided")
		}
		if err != nil {
			err := fmt.Errorf("failed to get secret name for IDPortenClient: %w", err)
			return nil, err
		}

		var notPaths *[]string
		if application.Spec.IDPorten.Authentication.IgnorePaths != nil {
			notPaths = application.Spec.IDPorten.Authentication.IgnorePaths
		} else {
			notPaths = nil
		}
		providerInfo = append(providerInfo, reconciliation.IdentityProviderInfo{
			Provider:   reconciliation.ID_PORTEN,
			SecretName: *secretName,
			NotPaths:   notPaths,
		})
	}
	if application.Spec.Maskinporten != nil && application.Spec.Maskinporten.Authentication != nil && application.Spec.Maskinporten.Authentication.Enabled == true {
		var secretName *string
		var err error
		if application.Spec.Maskinporten.Authentication.SecretName != nil {
			// If secret name is provided, use it regardless of whether Maskinporten is enabled
			secretName = application.Spec.Maskinporten.Authentication.SecretName
		} else if application.Spec.Maskinporten.Enabled {
			// If Maskinporten is enabled but no secretName provided, retrieve the generated secret from MaksinPortenClient
			secretName, err = getSecretNameForIdentityProvider(k8sClient, ctx,
				types.NamespacedName{
					Namespace: application.Namespace,
					Name:      application.Name,
				},
				reconciliation.MASKINPORTEN,
				application.UID)
		} else {
			// If Maskinporten is not enabled and no secretName provided, return error
			return nil, fmt.Errorf("JWT authentication requires either Maskinporten to be enabled or a secretName to be provided")
		}
		if err != nil {
			err := fmt.Errorf("failed to get secret name for MaskinPortenClient: %w", err)
			return nil, err
		}

		var notPaths *[]string
		if application.Spec.Maskinporten.Authentication.IgnorePaths != nil {
			notPaths = application.Spec.Maskinporten.Authentication.IgnorePaths
		} else {
			notPaths = nil
		}
		providerInfo = append(providerInfo, reconciliation.IdentityProviderInfo{
			Provider:   reconciliation.MASKINPORTEN,
			SecretName: *secretName,
			NotPaths:   notPaths,
		})
	}
	return &providerInfo, nil
}

func getSecretNameForIdentityProvider(k8sClient client.Client, ctx context.Context, namespacedName types.NamespacedName, provider reconciliation.IdentityProvider, applicationUID types.UID) (*string, error) {
	switch provider {
	case reconciliation.ID_PORTEN:
		idPortenClient, err := util.GetIdPortenClient(k8sClient, ctx, namespacedName)
		if err != nil {
			err := fmt.Errorf("failed to get IDPortenClient: %w", namespacedName.String())
			return nil, err
		}
		for _, ownerReference := range idPortenClient.OwnerReferences {
			if ownerReference.UID == applicationUID {
				return &idPortenClient.Spec.SecretName, nil
			}
		}
		err = fmt.Errorf("no IDPortenClient with ownerRef to (%w) found", namespacedName.String())
		return nil, err

	case reconciliation.MASKINPORTEN:
		maskinPortenClient, err := util.GetMaskinPortenlient(k8sClient, ctx, namespacedName)
		if err != nil {
			err := fmt.Errorf("failed to get MaskinPortenClient: %w", namespacedName.String())
			return nil, err
		}
		for _, ownerReference := range maskinPortenClient.OwnerReferences {
			if ownerReference.UID == applicationUID {
				return &maskinPortenClient.Spec.SecretName, nil
			}
		}
		err = fmt.Errorf("no MaskinPortenClient with ownerRef to (%w) found", namespacedName.String())
		return nil, err

	default:
		return nil, fmt.Errorf("provider: %w not supported", provider)
	}
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
			selector := metav1.LabelSelector{MatchLabels: rule.NamespacesByLabel}
			selectorString, err := metav1.LabelSelectorAsSelector(&selector)
			if err != nil {
				ruleErrors = append(ruleErrors, fmt.Errorf("failed to create label selector: %w", err))
			}
			namespaces := &corev1.NamespaceList{}
			if err = r.GetClient().List(ctx, namespaces, &client.ListOptions{LabelSelector: selectorString}); err != nil {
				ruleErrors = append(ruleErrors, fmt.Errorf("failed to list namespaces: %w", err))
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
