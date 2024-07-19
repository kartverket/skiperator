package controllers

import (
	"context"
	"fmt"
	"github.com/kartverket/skiperator/internal/controllers/common"
	"github.com/kartverket/skiperator/pkg/log"
	. "github.com/kartverket/skiperator/pkg/reconciliation"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/github"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/istio/sidecar"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/networkpolicy/defaultdeny"
	"github.com/kartverket/skiperator/pkg/util"
	istionetworkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type NamespaceReconciler struct {
	common.ReconcilerBase
	Token    string
	Registry string
}

//+kubebuilder:rbac:groups=core,resources=namespaces,verbs=get;list;watch
//+kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=networking.istio.io,resources=sidecars,verbs=get;list;watch;create;update;patch;delete

func (r *NamespaceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Namespace{}).
		Owns(&networkingv1.NetworkPolicy{}).
		Owns(&istionetworkingv1beta1.Sidecar{}).
		Owns(&corev1.Secret{}, builder.WithPredicates(
			util.MatchesPredicate[*corev1.Secret](github.IsImagePullSecret),
		)).
		Complete(r)
}

func (r *NamespaceReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	rLog := log.NewLogger().WithName(fmt.Sprintf("namespace: %s", req.Name))
	rLog.Debug("Starting reconcile for request", "requestName", req.Name)

	namespace := &corev1.Namespace{}
	err := r.GetClient().Get(ctx, req.NamespacedName, namespace)
	if errors.IsNotFound(err) {
		return common.DoNotRequeue()
	} else if err != nil {
		rLog.Error(err, "something went wrong fetching the namespace")
		r.EmitWarningEvent(namespace, "ReconcileStartFail", "something went wrong fetching the namespace, it might have been deleted")
		return common.RequeueWithError(err)
	}

	if r.isExcludedNamespace(ctx, namespace.Name) {
		rLog.Debug("Namespace is excluded from reconciliation", "name", namespace.Name)
		return common.RequeueWithError(err)
	}

	istioEnabled := r.IsIstioEnabledForNamespace(ctx, namespace.Name)
	identityConfigMap, err := r.GetIdentityConfigMap(ctx)
	if err != nil {
		rLog.Error(err, "cant find identity config map")
	}

	rLog.Debug("Starting reconciliation", "namespace", namespace.Name)
	r.EmitNormalEvent(namespace, "ReconcileStart", fmt.Sprintf("Namespace %v has started reconciliation loop", namespace.Name))

	reconciliation := NewNamespaceReconciliation(ctx, namespace, rLog, istioEnabled, r.GetRestConfig(), identityConfigMap)

	if err = defaultdeny.Generate(reconciliation); err != nil {
		return common.RequeueWithError(err)
	}
	//TODO if we can fix the constructor for github then we can do this in a nicer way
	if err = github.Generate(reconciliation, r.Token, r.Registry); err != nil {
		return common.RequeueWithError(err)
	}
	if err = sidecar.Generate(reconciliation); err != nil {
		return common.RequeueWithError(err)
	}

	if err = r.GetProcessor().Process(reconciliation); err != nil {
		return common.RequeueWithError(err)
	}

	r.EmitNormalEvent(namespace, "ReconcileEnd", fmt.Sprintf("Namespace %v has finished reconciliation loop", namespace.Name))

	return common.DoNotRequeue()
}

func (r *NamespaceReconciler) isExcludedNamespace(ctx context.Context, namespace string) bool {
	configMapNamespacedName := types.NamespacedName{Namespace: "skiperator-system", Name: "namespace-exclusions"}

	namespaceExclusionCMap, err := util.GetConfigMap(r.GetClient(), ctx, configMapNamespacedName)
	if err != nil {
		util.ErrDoPanic(err, "Something went wrong getting namespace-exclusion config map: %v")
	}

	nameSpacesToExclude := namespaceExclusionCMap.Data

	exclusion, keyExists := nameSpacesToExclude[namespace]

	return (keyExists && exclusion == "true")
}