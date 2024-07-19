package controllers

import (
	"context"
	"fmt"
	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/internal/controllers/common"
	"github.com/kartverket/skiperator/pkg/log"
	. "github.com/kartverket/skiperator/pkg/reconciliation"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/certificate"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/deployment"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/gcp/auth"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/hpa"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/idporten"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/istio/authorizationpolicy"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/istio/gateway"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/istio/peerauthentication"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/istio/serviceentry"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/istio/virtualservice"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/maskinporten"
	networkpolicy "github.com/kartverket/skiperator/pkg/resourcegenerator/networkpolicy/dynamic"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/pdb"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/resourceutils"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/service"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/serviceaccount"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/servicemonitor"
	"github.com/kartverket/skiperator/pkg/util"
	nais_io_v1 "github.com/nais/liberator/pkg/apis/nais.io/v1"
	pov1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"golang.org/x/exp/maps"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	policyv1 "k8s.io/api/policy/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// +kubebuilder:rbac:groups=skiperator.kartverket.no,resources=applications;applications/status,verbs=get;list;watch;update
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=events,verbs=create;patch
// +kubebuilder:rbac:groups=core,resources=services;configmaps;serviceaccounts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=policy,resources=poddisruptionbudgets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=networking.k8s.io,resources=networkpolicies,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=networking.istio.io,resources=gateways;serviceentries;virtualservices,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=security.istio.io,resources=peerauthentications;authorizationpolicies,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=autoscaling,resources=horizontalpodautoscalers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cert-manager.io,resources=certificates,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apiextensions.k8s.io,resources=customresourcedefinitions,verbs=get
// +kubebuilder:rbac:groups=monitoring.coreos.com,resources=servicemonitors,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=monitoring.coreos.com,resources=podmonitors,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=nais.io,resources=maskinportenclients;idportenclients,verbs=get;list;watch;create;update;patch;delete

type ApplicationReconciler struct {
	common.ReconcilerBase
}

const applicationFinalizer = "skip.statkart.no/finalizer"

func (r *ApplicationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&skiperatorv1alpha1.Application{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&networkingv1beta1.ServiceEntry{}).
		Owns(&networkingv1beta1.Gateway{}).
		Owns(&autoscalingv2.HorizontalPodAutoscaler{}).
		Owns(&networkingv1beta1.VirtualService{}).
		Owns(&securityv1beta1.PeerAuthentication{}).
		Owns(&corev1.ServiceAccount{}).
		Owns(&policyv1.PodDisruptionBudget{}).
		Owns(&networkingv1.NetworkPolicy{}).
		Owns(&securityv1beta1.AuthorizationPolicy{}).
		Owns(&nais_io_v1.MaskinportenClient{}).
		Owns(&nais_io_v1.IDPortenClient{}).
		Owns(&pov1.ServiceMonitor{}).
		Watches(&certmanagerv1.Certificate{}, handler.EnqueueRequestsFromMapFunc(handleApplicationCertRequest)).
		WithEventFilter(predicate.Or(predicate.GenerationChangedPredicate{}, predicate.LabelChangedPredicate{})).
		Complete(r)
}

type reconciliationFunc func(reconciliation Reconciliation) error

// TODO Clean up logs, events

func (r *ApplicationReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	rLog := log.NewLogger().WithName(fmt.Sprintf("application: %s", req.Name))
	rLog.Info("Starting reconcile", "request", req.Name)

	rdy := r.isClusterReady(ctx)
	if !rdy {
		panic("Cluster is not ready, missing servicemonitors.monitoring.coreos.com most likely")
	}

	application, err := r.getApplication(req, ctx)
	if application == nil {
		return ctrl.Result{}, nil
	} else if err != nil {
		return ctrl.Result{}, err
	}

	isApplicationMarkedToBeDeleted := application.GetDeletionTimestamp() != nil
	if isApplicationMarkedToBeDeleted {
		if err = r.finalizeApplication(application, ctx); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	if !common.ShouldReconcile(application) {
		return ctrl.Result{}, nil
	}

	// Copy application so we can check for diffs. Should be none on existing applications.
	tmpApplication := application.DeepCopy()

	r.setApplicationDefaults(application, ctx)

	specDiff, err := util.GetObjectDiff(tmpApplication.Spec, application.Spec)
	if err != nil {
		return common.RequeueWithError(err)
	}

	statusDiff, err := util.GetObjectDiff(tmpApplication.Status, application.Status)
	if err != nil {
		return common.RequeueWithError(err)
	}

	// If we update the Application initially on applied defaults before starting reconciling resources we allow all
	// updates to be visible even though the controllerDuties may take some time.
	if len(statusDiff) > 0 {
		rLog.Debug("Queueing for status diff")
		err := r.GetClient().Status().Update(ctx, application)
		return reconcile.Result{Requeue: true}, err
	}

	// Finalizer check is due to a bug when updating using controller-runtime
	// See https://github.com/kubernetes-sigs/controller-runtime/issues/2453
	if len(specDiff) > 0 || (!ctrlutil.ContainsFinalizer(tmpApplication, applicationFinalizer) && ctrlutil.ContainsFinalizer(application, applicationFinalizer)) {
		rLog.Debug("Queuing for spec diff")
		err := r.GetClient().Update(ctx, application)
		return reconcile.Result{Requeue: true}, err
	}

	istioEnabled := r.IsIstioEnabledForNamespace(ctx, application.Namespace)
	identityConfigMap, err := r.GetIdentityConfigMap(ctx)
	if err != nil {
		rLog.Error(err, "cant find identity config map")
	}

	//Start the actual reconciliation
	rLog.Debug("Starting reconciliation loop", "application", application.Name)
	r.GetRecorder().Eventf(
		application,
		"Normal",
		"ReconcileStart",
		fmt.Sprintf("Application %v has started reconciliation loop", application.Name))

	reconciliation := NewApplicationReconciliation(ctx, application, rLog, istioEnabled, r.GetRestConfig(), identityConfigMap)

	//TODO status and conditions in application object
	funcs := []reconciliationFunc{
		certificate.Generate,
		service.Generate,
		auth.Generate,
		serviceentry.Generate,
		gateway.Generate,
		virtualservice.Generate,
		hpa.Generate,
		peerauthentication.Generate,
		serviceaccount.Generate,
		networkpolicy.Generate,
		authorizationpolicy.Generate,
		pdb.Generate,
		servicemonitor.Generate,
		idporten.Generate,
		maskinporten.Generate,
		deployment.Generate,
	}

	for _, f := range funcs {
		if err := f(reconciliation); err != nil {
			rLog.Error(err, "Failed to create resource for application")
			return common.RequeueWithError(err)
		}
	}

	if err = r.setApplicationResourcesDefaults(reconciliation.GetResources(), application); err != nil {
		rLog.Error(err, "Failed to set application resource defaults")
		return common.RequeueWithError(err)
	}

	if err = r.GetProcessor().Process(reconciliation); err != nil {
		return common.RequeueWithError(err)
	}

	r.GetClient().Status().Update(ctx, application)

	return common.DoNotRequeue()
}

func (r *ApplicationReconciler) getApplication(req reconcile.Request, ctx context.Context) (*skiperatorv1alpha1.Application, error) {
	application := &skiperatorv1alpha1.Application{}
	if err := r.GetClient().Get(ctx, req.NamespacedName, application); err != nil {
		if errors.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("error when trying to get application: %w", err)
	}

	return application, nil
}

func (r *ApplicationReconciler) finalizeApplication(application *skiperatorv1alpha1.Application, ctx context.Context) error {
	if ctrlutil.ContainsFinalizer(application, applicationFinalizer) {
		ctrlutil.RemoveFinalizer(application, applicationFinalizer)
		err := r.GetClient().Update(ctx, application)
		if err != nil {
			return fmt.Errorf("Something went wrong when trying to finalize application. %w", err)
		}
	}

	return nil
}

func (r *ApplicationReconciler) setApplicationResourcesDefaults(
	resources []*client.Object,
	app *skiperatorv1alpha1.Application,
) error {
	for _, resource := range resources {
		if err := resourceutils.AddGVK(r.GetScheme(), *resource); err != nil {
			return err
		}
		resourceutils.SetCommonAnnotations(*resource)
		resourceutils.SetApplicationLabels(*resource, app)
		if err := resourceutils.SetOwnerReference(app, *resource, r.GetScheme()); err != nil {
			return err
		}
	}
	return nil
}

/*
 * Set application defaults. For existing applications this shouldn't do anything
 */
func (r *ApplicationReconciler) setApplicationDefaults(application *skiperatorv1alpha1.Application, ctx context.Context) {
	application.FillDefaultsSpec()
	if !ctrlutil.ContainsFinalizer(application, applicationFinalizer) {
		ctrlutil.AddFinalizer(application, applicationFinalizer)
	}

	// Add labels to application
	//TODO can we skip a step here?
	if application.Labels == nil {
		application.Labels = make(map[string]string)
	}
	maps.Copy(application.Labels, resourceutils.GetApplicationDefaultLabels(application))
	maps.Copy(application.Labels, application.Spec.Labels)

	// Add team label
	if len(application.Spec.Team) == 0 {
		if name, err := r.teamNameForNamespace(ctx, application); err == nil {
			application.Spec.Team = name
		}
	}

	application.FillDefaultsStatus()
}

func (r *ApplicationReconciler) isClusterReady(ctx context.Context) bool {
	if !r.isCrdPresent(ctx, "servicemonitors.monitoring.coreos.com") {
		return false
	}
	return true
}

func (r *ApplicationReconciler) teamNameForNamespace(ctx context.Context, app *skiperatorv1alpha1.Application) (string, error) {
	ns := &corev1.Namespace{}
	if err := r.GetClient().Get(ctx, types.NamespacedName{Name: app.Namespace}, ns); err != nil {
		return "", err
	}

	teamValue := ns.Labels["team"]
	if len(teamValue) > 0 {
		return teamValue, nil
	}
	return "", fmt.Errorf("missing value for team label")
}

// Name in the form of "servicemonitors.monitoring.coreos.com".
func (r *ApplicationReconciler) isCrdPresent(ctx context.Context, name string) bool {
	result, err := r.GetApiExtensionsClient().ApiextensionsV1().CustomResourceDefinitions().Get(ctx, name, metav1.GetOptions{})
	if err != nil || result == nil {
		return false
	}

	return true
}

func handleApplicationCertRequest(_ context.Context, obj client.Object) []reconcile.Request {
	cert, ok := obj.(*certmanagerv1.Certificate)
	if !ok {
		return nil
	}

	isSkiperatorOwned := cert.Labels["app.kubernetes.io/managed-by"] == "skiperator" &&
		cert.Labels["skiperator.skiperator.no/controller"] == "application"

	requests := make([]reconcile.Request, 0)

	if isSkiperatorOwned {
		requests = append(requests, reconcile.Request{
			NamespacedName: types.NamespacedName{
				Namespace: cert.Labels["application.skiperator.no/app-namespace"],
				Name:      cert.Labels["application.skiperator.no/app-name"],
			},
		})
	}

	return requests
}
