package controller

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/kartverket/skiperator/controllers/application"
	"github.com/kartverket/skiperator/pkg/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"

	policyv1 "k8s.io/api/policy/v1"

	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	nais_io_v1 "github.com/nais/liberator/pkg/apis/nais.io/v1"
	"golang.org/x/exp/maps"

	"github.com/kartverket/skiperator/pkg/util"
	pov1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
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
	client   client.Client
	recorder record.EventRecorder
}

const applicationFinalizer = "skip.statkart.no/finalizer"

func (r *ApplicationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&skiperatorv1alpha1.Application{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&networkingv1beta1.ServiceEntry{}).
		Owns(&networkingv1beta1.Gateway{}, builder.WithPredicates(
			util.MatchesPredicate[*networkingv1beta1.Gateway](applicationcontroller.isIngressGateway),
		)).
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
		Owns(&certmanagerv1.Certificate{}).
		WithEventFilter(predicate.Or(predicate.GenerationChangedPredicate{}, predicate.LabelChangedPredicate{})).
		Complete(r)
}

func (r *ApplicationReconciler) getApplication(req reconcile.Request, ctx context.Context) (*skiperatorv1alpha1.Application, error) {
	ctxLog := log.FromContext(ctx)
	ctxLog.Debug("Trying to get application from request", req)

	application := &skiperatorv1alpha1.Application{}
	if err := r.client.Get(ctx, req.NamespacedName, application); err != nil {
		if errors.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("error when trying to get application: %w", err)
	}

	return application, nil
}

func (r *ApplicationReconciler) finalizeApplication(application *skiperatorv1alpha1.Application, ctx context.Context) error {
	ctxLog := log.FromContext(ctx)
	ctxLog.Debug("finalizing application", application)

	if ctrlutil.ContainsFinalizer(application, applicationFinalizer) {
		ctrlutil.RemoveFinalizer(application, applicationFinalizer)
		err := r.client.Update(ctx, application)
		if err != nil {
			ctxLog.Error(err, "Something went wrong when trying to finalize application.")
			return err
		}
	}

	return nil
}

func shouldReconcile(application *skiperatorv1alpha1.Application) bool {
	labels := application.GetLabels()
	return labels["skiperator.kartverket.no/ignore"] != "true"
}

/*
 * Set application defaults. For existing applications this shouldn't do anything
 */
func (r *ApplicationReconciler) setApplicationDefaults(application *skiperatorv1alpha1.Application, ctx context.Context) {
	ctxLog := log.NewLogger(ctx)
	ctxLog.Debug("Setting application defaults", application.Name)

	application.FillDefaultsSpec()
	if !ctrlutil.ContainsFinalizer(application, applicationFinalizer) {
		ctrlutil.AddFinalizer(application, applicationFinalizer)
	}

	// Add labels to application
	application.Labels["app.kubernetes.io/managed-by"] = "skiperator"
	application.Labels["skiperator.skiperator.no/controller"] = "application"
	application.Labels["application.skiperator.no/app-name"] = application.Name
	application.Labels["application.skiperator.no/app-namespace"] = application.Namespace
	maps.Copy(application.Labels, application.Spec.Labels)

	// Add team label
	if len(application.Spec.Team) == 0 {
		if name, err := r.teamNameForNamespace(ctx, application); err == nil {
			application.Spec.Team = name
		}
	}

	application.FillDefaultsStatus()
}

func (r *ApplicationReconciler) isIstioEnabledInNamespace(ctx context.Context, namespaceName string) bool {
	namespace := corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespaceName,
		},
	}

	err := r.client.Get(ctx, client.ObjectKeyFromObject(&namespace), &namespace)
	if err != nil {
		return false
	}

	v, exists := namespace.Labels[util.IstioRevisionLabel]

	return exists && len(v) > 0
}

type Reconciliation interface {
	SomeFunctions()
}

type ReconciliationTask struct {
	ctx         context.Context
	application *skiperatorv1alpha1.Application
	logger      logr.Logger
	objects     []client.Object
}

func (r *ApplicationReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	ctxLog := log.NewLogger(ctx)
	ctxLog.Debug("Starting reconcile for request", req.Name)

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

	if !shouldReconcile(application) {
		return ctrl.Result{}, nil
	}

	// Copy application so we can check for diffs. Should be none on existing applications.
	tmpApplication := application.DeepCopy()

	r.setApplicationDefaults(application, ctx)

	specDiff, err := util.GetObjectDiff(tmpApplication.Spec, application.Spec)
	if err != nil {
		return util.RequeueWithError(err)
	}

	statusDiff, err := util.GetObjectDiff(tmpApplication.Status, application.Status)
	if err != nil {
		return util.RequeueWithError(err)
	}

	// If we update the Application initially on applied defaults before starting reconciling resources we allow all
	// updates to be visible even though the controllerDuties may take some time.
	if len(statusDiff) > 0 {
		err := r.client.Status().Update(ctx, application)
		return reconcile.Result{Requeue: true}, err
	}

	// Finalizer check is due to a bug when updating using controller-runtime
	// See https://github.com/kubernetes-sigs/controller-runtime/issues/2453
	if len(specDiff) > 0 || (!ctrlutil.ContainsFinalizer(tmpApplication, applicationFinalizer) && ctrlutil.ContainsFinalizer(application, applicationFinalizer)) {
		err := r.client.Update(ctx, application)
		return reconcile.Result{Requeue: true}, err
	}

	ctxLog.Debug("Starting reconciliation loop", application.Name)
	r.recorder.Eventf(
		application,
		"Normal",
		"ReconcileStart",
		fmt.Sprintf("Application %v has started reconciliation loop", application.Name))

	controllerDuties := []func(context.Context, *skiperatorv1alpha1.Application) (reconcile.Result, error){
		r.reconcilePodDisruptionBudget,
		r.reconcileServiceMonitor,
		r.reconcileIDPorten,
		r.reconcileMaskinporten,
		r.reconcileDeployment,
	}

	for _, fn := range controllerDuties {
		res, err := fn(ctx, application)
		if err != nil {
			r.GetClient().Status().Update(ctx, application)
			return res, err
		} else if res.RequeueAfter > 0 || res.Requeue {
			r.GetClient().Status().Update(ctx, application)
			return res, nil
		}
	}
	r.GetClient().Status().Update(ctx, application)
	r.EmitNormalEvent(application, "ReconcileEnd", fmt.Sprintf("Application %v has finished reconciliation loop", application.Name))

	return util.RequeueWithError(err)
}

func (r *ApplicationReconciler) teamNameForNamespace(ctx context.Context, app *skiperatorv1alpha1.Application) (string, error) {
	ctxLog := log.FromContext(ctx)
	ctxLog.Debug("Trying to get team name for namespace", app.Namespace)
	ns := &corev1.Namespace{}
	if err := r.client.Get(ctx, types.NamespacedName{Name: app.Namespace}, ns); err != nil {
		return "", err
	}

	teamValue := ns.Labels["team"]
	if len(teamValue) > 0 {
		return teamValue, nil
	}
	ctxLog.Warning("Missing value for team label in namespace", app.Namespace, app.Name)
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

func (r *ApplicationReconciler) manageControllerStatus(context context.Context, app *skiperatorv1alpha1.Application, controller string, statusName skiperatorv1alpha1.StatusNames, message string) (reconcile.Result, error) {
	app.UpdateControllerStatus(controller, message, statusName)
	return util.DoNotRequeue()
}

func (r *ApplicationReconciler) manageControllerStatusError(context context.Context, app *skiperatorv1alpha1.Application, controller string, issue error) (reconcile.Result, error) {
	app.UpdateControllerStatus(controller, issue.Error(), skiperatorv1alpha1.ERROR)
	r.EmitWarningEvent(app, "ControllerFault", fmt.Sprintf("%v controller experienced an error: %v", controller, issue.Error()))
	return util.RequeueWithError(issue)
}

func (r *ApplicationReconciler) SetControllerPending(context context.Context, app *skiperatorv1alpha1.Application, controller string) (reconcile.Result, error) {
	message := controller + " has been initialized and is pending Skiperator startup"

	return r.manageControllerStatus(context, app, controller, skiperatorv1alpha1.PENDING, message)
}

func (r *ApplicationReconciler) SetControllerProgressing(context context.Context, app *skiperatorv1alpha1.Application, controller string) (reconcile.Result, error) {
	message := controller + " has started sync"

	return r.manageControllerStatus(context, app, controller, skiperatorv1alpha1.PROGRESSING, message)
}

func (r *ApplicationReconciler) SetControllerSynced(context context.Context, app *skiperatorv1alpha1.Application, controller string) (reconcile.Result, error) {
	message := controller + " has finished synchronizing"

	return r.manageControllerStatus(context, app, controller, skiperatorv1alpha1.SYNCED, message)
}

func (r *ApplicationReconciler) SetControllerError(context context.Context, app *skiperatorv1alpha1.Application, controller string, issue error) (reconcile.Result, error) {
	return r.manageControllerStatusError(context, app, controller, issue)
}

func (r *ApplicationReconciler) SetControllerFinishedOutcome(context context.Context, app *skiperatorv1alpha1.Application, controllerName string, issue error) (reconcile.Result, error) {
	if issue != nil {
		return r.manageControllerStatusError(context, app, controllerName, issue)
	}

	return r.SetControllerSynced(context, app, controllerName)
}

func (r *ApplicationReconciler) getGCPIdentityConfigMap(ctx context.Context, application skiperatorv1alpha1.Application) (*corev1.ConfigMap, error) {
	if skipJob.Spec.Container.GCP != nil {
		gcpIdentityConfigMapNamespacedName := types.NamespacedName{Namespace: "skiperator-system", Name: "gcp-identity-config"}

		configMap, err := util.GetConfigMap(r.GetClient(), ctx, gcpIdentityConfigMapNamespacedName)
		if !util.ErrIsMissingOrNil(
			r.GetRecorder(),
			err,
			"Cannot find configmap named "+gcpIdentityConfigMapNamespacedName.Name+" in namespace "+gcpIdentityConfigMapNamespacedName.Namespace,
			&skipJob,
		) {
			return nil, err
		}

		return &configMap, nil
	} else {
		return nil, nil
	}
}
