package controllers

import (
	"context"
	"fmt"
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
	"github.com/kartverket/skiperator/pkg/resourcegenerator/serviceaccount"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/servicemonitor"
	"github.com/kartverket/skiperator/pkg/resourceprocessor"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
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
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
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
	client           client.Client
	recorder         record.EventRecorder
	restConfig       *rest.Config
	processor        *resourceprocessor.ResourceProcessor
	extensionsClient *apiextensionsclient.Clientset
}

const applicationFinalizer = "skip.statkart.no/finalizer"

func NewApplicationReconciler(
	client client.Client,
	restConfig *rest.Config,
	recorder record.EventRecorder,
	processor *resourceprocessor.ResourceProcessor,
	extensionsClient *apiextensionsclient.Clientset,
) *ApplicationReconciler {
	return &ApplicationReconciler{
		client:           client,
		recorder:         recorder,
		restConfig:       restConfig,
		processor:        processor,
		extensionsClient: extensionsClient,
	}
}

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
		Owns(&certmanagerv1.Certificate{}).
		Complete(r)
}

type reconciliationFunc func(reconciliation Reconciliation) error

func (r *ApplicationReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	ctxLog := log.NewLogger(ctx).WithName("application-controller")
	ctxLog.Debug("Starting reconcile", "request", req.Name)

	rdy := r.isClusterReady(ctx)
	if !rdy {
		ctxLog.Warning("Cluster is not ready for reconciliation", "request", req.Name)
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

	identityConfigMap, err := getIdentityConfigMap(r.client)
	if err != nil {
		ctxLog.Error(err, "cant find identity config map")
	}

	//Start the actual reconciliation
	ctxLog.Debug("Starting reconciliation loop", "application", application.Name)
	r.recorder.Eventf(
		application,
		"Normal",
		"ReconcileStart",
		fmt.Sprintf("Application %v has started reconciliation loop", application.Name))

	reconciliation := NewApplicationReconciliation(ctx, application, ctxLog, r.restConfig, identityConfigMap)

	//TODO status and conditions in application object
	funcs := []reconciliationFunc{
		certificate.Generate,
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
			return util.RequeueWithError(err)
		}
	}

	if err = r.processor.Process(reconciliation); err != nil {
		return util.RequeueWithError(err)
	}

	r.client.Status().Update(ctx, application)

	return util.DoNotRequeue()
}

func (r *ApplicationReconciler) getApplication(req reconcile.Request, ctx context.Context) (*skiperatorv1alpha1.Application, error) {
	ctxLog := log.FromContext(ctx)
	ctxLog.Debug("Trying to get application from request", "request", req)

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
	ctxLog.Debug("finalizing application", "application", application)

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

func getApplicationDefaultLabels(application *skiperatorv1alpha1.Application) map[string]string {
	return map[string]string{
		"app.kubernetes.io/managed-by":            "skiperator",
		"skiperator.skiperator.no/controller":     "application",
		"application.skiperator.no/app-name":      application.Name,
		"application.skiperator.no/app-namespace": application.Namespace,
	}
}

/*
 * Set application defaults. For existing applications this shouldn't do anything
 */
func (r *ApplicationReconciler) setApplicationDefaults(application *skiperatorv1alpha1.Application, ctx context.Context) {
	ctxLog := log.NewLogger(ctx)
	ctxLog.Debug("Setting application defaults", "application", application.Name)

	application.FillDefaultsSpec()
	if !ctrlutil.ContainsFinalizer(application, applicationFinalizer) {
		ctrlutil.AddFinalizer(application, applicationFinalizer)
	}

	// Add labels to application
	//TODO can we skip a step here?
	maps.Copy(getApplicationDefaultLabels(application), application.Labels)
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

func (r *ApplicationReconciler) isClusterReady(ctx context.Context) bool {
	if !r.isCrdPresent(ctx, "servicemonitors.monitoring.coreos.com") {
		return false
	}
	return true
}

func (r *ApplicationReconciler) teamNameForNamespace(ctx context.Context, app *skiperatorv1alpha1.Application) (string, error) {
	ctxLog := log.FromContext(ctx)
	ctxLog.Debug("Trying to get team name for namespace", "namespace", app.Namespace)
	ns := &corev1.Namespace{}
	if err := r.client.Get(ctx, types.NamespacedName{Name: app.Namespace}, ns); err != nil {
		return "", err
	}

	teamValue := ns.Labels["team"]
	if len(teamValue) > 0 {
		return teamValue, nil
	}
	ctxLog.Warning("Missing value for team label in namespace", "namespace", app.Namespace, "applicationName", app.Name)
	return "", fmt.Errorf("missing value for team label")
}

// Name in the form of "servicemonitors.monitoring.coreos.com".
func (r *ApplicationReconciler) isCrdPresent(ctx context.Context, name string) bool {
	result, err := r.extensionsClient.ApiextensionsV1().CustomResourceDefinitions().Get(ctx, name, metav1.GetOptions{})
	if err != nil || result == nil {
		return false
	}

	return true
}
