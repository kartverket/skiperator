package controllers

import (
	"context"
	"fmt"
	"github.com/kartverket/skiperator/api/v1alpha1/digdirator"
	jwtAuth "github.com/kartverket/skiperator/pkg/auth"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/idporten"
	allowAuthPolicy "github.com/kartverket/skiperator/pkg/resourcegenerator/istio/authorizationpolicy/allow"
	denyAuthPolicy "github.com/kartverket/skiperator/pkg/resourcegenerator/istio/authorizationpolicy/default_deny"
	jwtAuthPolicy "github.com/kartverket/skiperator/pkg/resourcegenerator/istio/authorizationpolicy/jwt_auth"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/maskinporten"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/pdb"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/prometheus"
	"regexp"
	"strings"

	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/internal/controllers/common"
	"github.com/kartverket/skiperator/pkg/log"
	. "github.com/kartverket/skiperator/pkg/reconciliation"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/certificate"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/deployment"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/gcp/auth"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/hpa"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/istio/gateway"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/istio/peerauthentication"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/istio/requestauthentication"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/istio/serviceentry"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/istio/telemetry"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/istio/virtualservice"
	networkpolicy "github.com/kartverket/skiperator/pkg/resourcegenerator/networkpolicy/dynamic"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/resourceutils"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/service"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/serviceaccount"
	"github.com/kartverket/skiperator/pkg/util"
	nais_io_v1 "github.com/nais/liberator/pkg/apis/nais.io/v1"
	pov1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"golang.org/x/exp/maps"
	istionetworkingv1 "istio.io/client-go/pkg/apis/networking/v1"
	securityv1 "istio.io/client-go/pkg/apis/security/v1"
	telemetryv1 "istio.io/client-go/pkg/apis/telemetry/v1"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	policyv1 "k8s.io/api/policy/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
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
// +kubebuilder:rbac:groups=telemetry.istio.io,resources=telemetries,verbs=get;list;watch;create;update;patch;delete
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

var hostMatchExpression = regexp.MustCompile(`^([a-z0-9]+(-[a-z0-9]+)*\.)+[a-z]{2,}$`)

// TODO Watch applications that are using dynamic port allocation
func (r *ApplicationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&skiperatorv1alpha1.Application{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&istionetworkingv1.ServiceEntry{}).
		Owns(&istionetworkingv1.Gateway{}, builder.WithPredicates(
			util.MatchesPredicate[*istionetworkingv1.Gateway](isIngressGateway),
		)).
		Owns(&telemetryv1.Telemetry{}).
		Owns(&autoscalingv2.HorizontalPodAutoscaler{}).
		Owns(&istionetworkingv1.VirtualService{}).
		Owns(&securityv1.PeerAuthentication{}).
		Owns(&corev1.ServiceAccount{}).
		Owns(&policyv1.PodDisruptionBudget{}).
		Owns(&networkingv1.NetworkPolicy{}).
		Owns(&securityv1.RequestAuthentication{}).
		Owns(&securityv1.AuthorizationPolicy{}).
		Owns(&nais_io_v1.MaskinportenClient{}).
		Owns(&nais_io_v1.IDPortenClient{}).
		Owns(&pov1.ServiceMonitor{}).
		Watches(&corev1.Secret{}, handler.EnqueueRequestsFromMapFunc(handleDigdiratorSecret)).
		Watches(&certmanagerv1.Certificate{}, handler.EnqueueRequestsFromMapFunc(handleApplicationCertRequest)).
		WithEventFilter(predicate.Or(predicate.GenerationChangedPredicate{}, predicate.LabelChangedPredicate{})).
		Complete(r)
}

type reconciliationFunc func(reconciliation Reconciliation) error

// TODO Clean up logs, events

func (r *ApplicationReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	rLog := log.NewLogger().WithName(fmt.Sprintf("application-controller: %s", req.Name))
	rLog.Info("Starting reconcile", "request", req.Name)

	rdy := r.isClusterReady(ctx)
	if !rdy {
		panic("Cluster is not ready, missing servicemonitors.monitoring.coreos.com most likely")
	}

	application, err := r.getApplication(ctx, req)
	if application == nil {
		rLog.Info("Application not found, cleaning up watched resources", "application", req.Name)
		if errs := r.cleanUpWatchedResources(ctx, req.NamespacedName); len(errs) > 0 {
			return common.RequeueWithError(fmt.Errorf("error when trying to clean up watched resources: %w", errs[0]))
		}
		return common.DoNotRequeue()
	} else if err != nil {
		return common.RequeueWithError(err)
	}

	// TODO do we need this actually?
	isApplicationMarkedToBeDeleted := application.GetDeletionTimestamp() != nil
	if isApplicationMarkedToBeDeleted {
		if err = r.finalizeApplication(application, ctx); err != nil {
			return ctrl.Result{}, err
		}
		return common.DoNotRequeue()
	}

	if !common.ShouldReconcile(application) {
		return common.DoNotRequeue()
	}

	if err := validateIngresses(application); err != nil {
		rLog.Error(err, "invalid ingress in application manifest")
		r.SetErrorState(ctx, application, err, "invalid ingress in application manifest", "InvalidApplication")
		return common.RequeueWithError(err)
	}

	authConfigs, err := r.getAuthConfigsForApplication(ctx, application)
	if err != nil {
		rLog.Error(err, "unable to resolve auth config for application", "application", application.Name)
	}

	// Copy application so we can check for diffs. Should be none on existing applications.
	tmpApplication := application.DeepCopy()

	r.setApplicationDefaults(application, ctx)

	specDiff, err := common.GetObjectDiff(tmpApplication.Spec, application.Spec)
	if err != nil {
		return common.RequeueWithError(err)
	}

	statusDiff, err := common.GetObjectDiff(tmpApplication.Status, application.Status)
	if err != nil {
		return common.RequeueWithError(err)
	}

	if len(statusDiff) > 0 {
		rLog.Info("Status has changed", "diff", statusDiff)
		err = r.GetClient().Status().Update(ctx, application)
		return reconcile.Result{Requeue: true}, err
	}

	// Finalizer check is due to a bug when updating using controller-runtime
	// See https://github.com/kubernetes-sigs/controller-runtime/issues/2453
	if len(specDiff) > 0 || (!ctrlutil.ContainsFinalizer(tmpApplication, applicationFinalizer) && ctrlutil.ContainsFinalizer(application, applicationFinalizer)) {
		rLog.Debug("Queuing for spec diff")
		err := r.GetClient().Update(ctx, application)
		return reconcile.Result{Requeue: true}, err
	}

	//We try to feed the access policy with port values dynamically,
	//if unsuccessfull we just don't set ports, and rely on podselectors
	r.UpdateAccessPolicy(ctx, application)

	//Start the actual reconciliation
	rLog.Debug("Starting reconciliation loop", "application", application.Name)
	r.SetProgressingState(ctx, application, fmt.Sprintf("Application %v has started reconciliation loop", application.Name))

	istioEnabled := r.IsIstioEnabledForNamespace(ctx, application.Namespace)
	identityConfigMap, err := r.GetIdentityConfigMap(ctx)
	if err != nil {
		rLog.Error(err, "cant find identity config map")
	} //TODO Error state?

	reconciliation := NewApplicationReconciliation(ctx, application, rLog, istioEnabled, r.GetRestConfig(), identityConfigMap, authConfigs)

	//TODO status and conditions in application object
	funcs := []reconciliationFunc{
		certificate.Generate,
		service.Generate,
		auth.Generate,
		serviceentry.Generate,
		gateway.Generate,
		virtualservice.Generate,
		telemetry.Generate,
		hpa.Generate,
		peerauthentication.Generate,
		serviceaccount.Generate,
		networkpolicy.Generate,
		denyAuthPolicy.Generate,
		allowAuthPolicy.Generate,
		jwtAuthPolicy.Generate,
		requestauthentication.Generate,
		pdb.Generate,
		prometheus.Generate,
		idporten.Generate,
		maskinporten.Generate,
		deployment.Generate,
	}

	for _, f := range funcs {
		if err = f(reconciliation); err != nil {
			rLog.Error(err, "failed to generate application resource")
			//At this point we don't have the gvk of the resource yet, so we can't set subresource status.
			r.SetErrorState(ctx, application, err, "failed to generate application resource", "ResourceGenerationFailure")
			return common.RequeueWithError(err)
		}
	}

	// We need to do this here, so we are sure it's done. Not setting GVK can cause big issues
	if err = r.setApplicationResourcesDefaults(reconciliation.GetResources(), application); err != nil {
		rLog.Error(err, "failed to set application resource defaults")
		r.SetErrorState(ctx, application, err, "failed to set application resource defaults", "ResourceDefaultsFailure")
		return common.RequeueWithError(err)
	}

	if errs := r.GetProcessor().Process(reconciliation); len(errs) > 0 {
		for _, err = range errs {
			rLog.Error(err, "failed to process resource")
			r.EmitWarningEvent(application, "ReconcileEndFail", fmt.Sprintf("Failed to process application resources: %s", err.Error()))
		}
		r.SetErrorState(ctx, application, fmt.Errorf("found %d errors", len(errs)), "failed to process application resources, see subresource status", "ProcessorFailure")
		return common.RequeueWithError(err)
	}

	r.updateConditions(application)
	r.SetSyncedState(ctx, application, "Application has been reconciled")

	return common.DoNotRequeue()
}

func (r *ApplicationReconciler) updateConditions(app *skiperatorv1alpha1.Application) {
	var conditions []metav1.Condition
	accessPolicy := app.Spec.AccessPolicy
	if accessPolicy != nil && !common.IsInternalRulesValid(accessPolicy) {
		conditions = append(conditions, common.GetInternalRulesCondition(app, metav1.ConditionFalse))
		app.Status.AccessPolicies = skiperatorv1alpha1.INVALIDCONFIG
	} else {
		conditions = append(conditions, common.GetInternalRulesCondition(app, metav1.ConditionTrue))
		app.Status.AccessPolicies = skiperatorv1alpha1.READY
	}
	app.Status.Conditions = conditions
}

func (r *ApplicationReconciler) getApplication(ctx context.Context, req reconcile.Request) (*skiperatorv1alpha1.Application, error) {
	application := &skiperatorv1alpha1.Application{}
	if err := r.GetClient().Get(ctx, req.NamespacedName, application); err != nil {
		if errors.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("error when trying to get application: %w", err)
	}

	return application, nil
}

func (r *ApplicationReconciler) cleanUpWatchedResources(ctx context.Context, name types.NamespacedName) []error {
	app := &skiperatorv1alpha1.Application{}
	app.SetName(name.Name)
	app.SetNamespace(name.Namespace)

	reconciliation := NewApplicationReconciliation(ctx, app, log.NewLogger(), false, nil, nil, nil)
	return r.GetProcessor().Process(reconciliation)
}

func (r *ApplicationReconciler) finalizeApplication(application *skiperatorv1alpha1.Application, ctx context.Context) error {
	if ctrlutil.ContainsFinalizer(application, applicationFinalizer) {
		ctrlutil.RemoveFinalizer(application, applicationFinalizer)
		err := r.GetClient().Update(ctx, application)
		if err != nil {
			return fmt.Errorf("something went wrong when trying to finalize application. %w", err)
		}
	}

	return nil
}

func (r *ApplicationReconciler) setApplicationResourcesDefaults(resources []client.Object, app *skiperatorv1alpha1.Application) error {
	for _, resource := range resources {
		if err := r.SetSubresourceDefaults(resources, app); err != nil {
			return err
		}
		resourceutils.SetApplicationLabels(resource, app)
	}

	//TODO should try to combine this with the above
	resourceLabelsWithNoMatch := resourceutils.FindResourceLabelErrors(app, resources)
	for k, _ := range resourceLabelsWithNoMatch {
		r.EmitWarningEvent(app, "MistypedLabel", fmt.Sprintf("Resource label %s not a generated resource", k))
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
	maps.Copy(application.Labels, application.GetDefaultLabels())
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

func handleDigdiratorSecret(_ context.Context, obj client.Object) []reconcile.Request {
	secret, ok := obj.(*corev1.Secret)
	if !ok {
		return nil
	}

	requests := make([]reconcile.Request, 0)

	// Check if secret is owned by digdirator with type digdirator.nais.io or maskinporten.digdirator.nais.io
	if strings.Contains(secret.Labels["type"], "digdirator.nais.io") {
		requests = append(requests, reconcile.Request{
			NamespacedName: types.NamespacedName{
				Namespace: secret.Namespace,
				Name:      secret.Labels["app"],
			},
		})
	}

	return requests
}

func handleApplicationCertRequest(_ context.Context, obj client.Object) []reconcile.Request {
	cert, ok := obj.(*certmanagerv1.Certificate)
	if !ok {
		return nil
	}

	isSkiperatorOwned := cert.Labels["app.kubernetes.io/managed-by"] == "skiperator" &&
		cert.Labels["skiperator.kartverket.no/controller"] == "application"

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

func isIngressGateway(gateway *istionetworkingv1.Gateway) bool {
	match, _ := regexp.MatchString("^.*-ingress-.*$", gateway.Name)

	return match
}

// TODO should be handled better
func validateIngresses(application *skiperatorv1alpha1.Application) error {
	var err error
	hosts, err := application.Spec.Hosts()
	if err != nil {
		return err
	}

	// TODO: Remove/rewrite?
	for _, h := range hosts.AllHosts() {
		if !hostMatchExpression.MatchString(h.Hostname) {
			errMessage := fmt.Sprintf("ingress with value '%s' was not valid. ingress must be lower case, contain no spaces, be a non-empty string, and have a hostname/domain separated by a period", h.Hostname)
			return errors.NewInvalid(application.GroupVersionKind().GroupKind(), application.Name, field.ErrorList{
				field.Invalid(field.NewPath("application").Child("spec").Child("ingresses"), application.Spec.Ingresses, errMessage),
			})
		}
	}
	return nil
}

func (r *ApplicationReconciler) getAuthConfigsForApplication(ctx context.Context, application *skiperatorv1alpha1.Application) (*jwtAuth.AuthConfigs, error) {
	var authConfigs jwtAuth.AuthConfigs

	providers := []digdirator.DigdiratorProvider{
		application.Spec.IDPorten,
		application.Spec.Maskinporten,
	}
	for _, provider := range providers {
		if provider.IsRequestAuthEnabled() {
			authConfig, err := r.getAuthConfig(ctx, *application, provider)
			if err != nil {
				return nil, fmt.Errorf("could not get auth config for provider '%s': %w", provider.GetDigdiratorName(), err)
			}
			authConfigs = append(authConfigs, *authConfig)
		}
	}
	if len(authConfigs) > 0 {
		authConfigs.IgnorePathsFromOtherAuthConfigs()
		return &authConfigs, nil
	} else {
		return nil, nil
	}
}

func (r *ApplicationReconciler) getAuthConfig(ctx context.Context, application skiperatorv1alpha1.Application, digdiratorProvider digdirator.DigdiratorProvider) (*jwtAuth.AuthConfig, error) {
	secret, err := r.getAuthConfigSecret(ctx, application, digdiratorProvider)
	if err != nil {
		return nil, fmt.Errorf("failed to get auth config secret for %s: %w", digdiratorProvider.GetDigdiratorName(), err)
	}
	return &jwtAuth.AuthConfig{
		Spec:        digdiratorProvider.GetAuthSpec(),
		Paths:       digdiratorProvider.GetPaths(),
		IgnorePaths: digdiratorProvider.GetIgnoredPaths(),
		ProviderURIs: digdirator.DigdiratorURIs{
			Name:      digdiratorProvider.GetDigdiratorName(),
			IssuerURI: string(secret.Data[digdiratorProvider.GetIssuerKey()]),
			JwksURI:   string(secret.Data[digdiratorProvider.GetJwksKey()]),
			ClientID:  string(secret.Data[digdiratorProvider.GetClientIDKey()]),
		},
	}, nil
}

func (r *ApplicationReconciler) getAuthConfigSecret(ctx context.Context, application skiperatorv1alpha1.Application, digdiratorProvider digdirator.DigdiratorProvider) (*corev1.Secret, error) {
	var secretName *string
	var err error

	if digdiratorProvider.GetProvidedSecretName() != nil {
		secretName = digdiratorProvider.GetProvidedSecretName()
	} else {
		secretName, err = r.getDigdiratorSecretName(ctx, digdiratorProvider, application)
		if err != nil {
			return nil, err
		}
	}

	namespacedName := types.NamespacedName{
		Name:      *secretName,
		Namespace: application.Namespace,
	}

	secret, err := util.GetSecret(r.GetClient(), ctx, namespacedName)
	if err != nil {
		return nil, err
	}

	return &secret, nil
}

func (r *ApplicationReconciler) getDigdiratorSecretName(ctx context.Context, digdiratorProvider digdirator.DigdiratorProvider, application skiperatorv1alpha1.Application) (*string, error) {
	var digdiratorClient digdirator.DigdiratorClient
	var err error

	namespacedName := types.NamespacedName{
		Name:      application.Name,
		Namespace: application.Namespace,
	}

	if digdiratorProvider.GetDigdiratorName() == digdirator.MaskinPortenName {
		digdiratorClient, err = util.GetMaskinportenClient(r.GetClient(), ctx, namespacedName)
		if err != nil {
			return nil, err
		}
	} else {
		digdiratorClient, err = util.GetIdPortenClient(r.GetClient(), ctx, namespacedName)
		if err != nil {
			return nil, err
		}
	}
	ownershipRefs := digdiratorClient.GetOwnerReferences()
	secretName := digdiratorClient.GetSecretName()

	for _, ownershipRef := range ownershipRefs {
		if ownershipRef.UID == application.UID {
			return &secretName, nil
		}
	}

	return nil, fmt.Errorf("digdirator client doesn't exist: %s", namespacedName)
}
