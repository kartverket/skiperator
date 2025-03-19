package controllers

import (
	"context"
	"fmt"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/internal/controllers/common"
	"github.com/kartverket/skiperator/pkg/log"
	. "github.com/kartverket/skiperator/pkg/reconciliation"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/imagepullsecret"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/istio/sidecar"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/networkpolicy/defaultdeny"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/resourceutils"
	"github.com/kartverket/skiperator/pkg/util"
	istionetworkingv1 "istio.io/client-go/pkg/apis/networking/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type NamespaceReconciler struct {
	common.ReconcilerBase
	PullSecret *imagepullsecret.ImagePullSecret
}

//+kubebuilder:rbac:groups=core,resources=namespaces,verbs=get;list;watch
//+kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=networking.istio.io,resources=sidecars,verbs=get;list;watch;create;update;patch;delete

func (r *NamespaceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Namespace{}).
		Owns(&networkingv1.NetworkPolicy{}).
		Owns(&istionetworkingv1.Sidecar{}).
		Owns(&corev1.Secret{}, builder.WithPredicates(
			util.MatchesPredicate[*corev1.Secret](imagepullsecret.IsImagePullSecret),
		)).
		Complete(r)
}

// TODO Move controller to argocd
func (r *NamespaceReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	rLog := log.NewLogger().WithName(fmt.Sprintf("namespace-controller: %s", req.Name))
	rLog.Debug("Starting reconcile for request", "requestName", req.Name)

	namespace := &corev1.Namespace{}
	err := r.GetClient().Get(ctx, req.NamespacedName, namespace)
	if errors.IsNotFound(err) || common.IsNamespaceTerminating(namespace) {
		rLog.Debug("Namespace is terminating or not found", "name", namespace.Name)
		return common.DoNotRequeue()
	} else if err != nil {
		rLog.Error(err, "something went wrong fetching the namespace")
		r.EmitWarningEvent(namespace, "ReconcileStartFail", "something went wrong fetching the namespace, it might have been deleted")
		return common.RequeueWithError(err)
	}

	if r.isExcludedNamespace(ctx, namespace.Name) {
		rLog.Debug("Namespace is excluded from reconciliation", "name", namespace.Name)
		return common.DoNotRequeue()
	}
	//This is a hack because namespace shouldn't be here. We need this to keep things generic
	SKIPNamespace := skiperatorv1alpha1.SKIPNamespace{Namespace: namespace}

	istioEnabled := r.IsIstioEnabledForNamespace(ctx, namespace.Name)
	identityConfigMap, err := r.GetIdentityConfigMap(ctx)
	if err != nil {
		rLog.Error(err, "cant find identity config map")
	}

	rLog.Debug("Starting reconciliation", "namespace", namespace.Name)
	r.EmitNormalEvent(namespace, "ReconcileStart", fmt.Sprintf("Namespace %v has started reconciliation loop", namespace.Name))
	reconciliation := NewNamespaceReconciliation(ctx, SKIPNamespace, rLog, istioEnabled, r.GetRestConfig(), identityConfigMap)

	funcs := []reconciliationFunc{
		sidecar.Generate,
		defaultdeny.Generate,
		r.PullSecret.Generate,
	}

	for _, f := range funcs {
		if err = f(reconciliation); err != nil {
			rLog.Error(err, "failed to generate namespace resource")
			return common.RequeueWithError(err)
		}
	}

	if err = r.setResourceDefaults(reconciliation.GetResources(), &SKIPNamespace); err != nil {
		rLog.Error(err, "Failed to set namespace resource defaults")
		r.EmitWarningEvent(namespace, "ReconcileEndFail", "Failed to set namespace resource defaults")
		return common.RequeueWithError(err)
	}

	if errs := r.GetProcessor().Process(reconciliation); len(errs) > 0 {
		rLog.Error(errs[0], "failed to process resources - returning only the first error", "numberOfErrors", len(errs))
		return common.RequeueWithError(errs[0])
	}

	r.EmitNormalEvent(namespace, "ReconcileEnd", fmt.Sprintf("Namespace %v has finished reconciliation loop", namespace.Name))

	return common.DoNotRequeue()
}

func (r *NamespaceReconciler) setResourceDefaults(resources []client.Object, skipns *skiperatorv1alpha1.SKIPNamespace) error {
	for _, resource := range resources {
		if err := resourceutils.AddGVK(r.GetScheme(), resource); err != nil {
			return err
		}
		resourceutils.SetNamespaceLabels(resource, skipns)
	}
	return nil
}

func (r *NamespaceReconciler) isExcludedNamespace(ctx context.Context, namespace string) bool {
	configMapNamespacedName := types.NamespacedName{Namespace: "skiperator-system", Name: "namespace-exclusions"}

	namespaceExclusionCMap, err := util.GetConfigMap(r.GetClient(), ctx, configMapNamespacedName)
	if err != nil {
		util.ErrDoPanic(err, "Something went wrong getting namespace-exclusion config map: %v")
	}

	nameSpacesToExclude := namespaceExclusionCMap.Data

	exclusion, keyExists := nameSpacesToExclude[namespace]

	return keyExists && exclusion == "true"
}
