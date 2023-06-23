package applicationcontroller

import (
	"context"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"regexp"

	policyv1 "k8s.io/api/policy/v1"

	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"golang.org/x/exp/maps"

	"github.com/kartverket/skiperator/pkg/util"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
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
// +kubebuilder:rbac:groups=security.istio.io,resources=peerauthentications;authorizationpolicies,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=autoscaling,resources=horizontalpodautoscalers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cert-manager.io,resources=certificates,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apiextensions.k8s.io,resources=customresourcedefinitions,verbs=get
// +kubebuilder:rbac:groups=monitoring.coreos.com,resources=servicemonitors,verbs=get;list;watch;create;update;patch;delete

type ApplicationReconciler struct {
	util.ReconcilerBase
}

const applicationFinalizer = "skip.statkart.no/finalizer"

func (r *ApplicationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&skiperatorv1alpha1.Application{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&networkingv1beta1.ServiceEntry{}, builder.WithPredicates(
			util.MatchesPredicate[*networkingv1beta1.ServiceEntry](isEgressServiceEntry),
		)).
		Owns(&networkingv1beta1.Gateway{}, builder.WithPredicates(
			util.MatchesPredicate[*networkingv1beta1.Gateway](isIngressGateway),
		)).
		Owns(&autoscalingv2.HorizontalPodAutoscaler{}).
		Owns(&networkingv1beta1.VirtualService{}).
		Owns(&securityv1beta1.PeerAuthentication{}).
		Owns(&corev1.ServiceAccount{}).
		Owns(&policyv1.PodDisruptionBudget{}).
		Owns(&networkingv1.NetworkPolicy{}).
		Owns(&securityv1beta1.AuthorizationPolicy{}).
		Watches(&certmanagerv1.Certificate{}, handler.EnqueueRequestsFromMapFunc(r.SkiperatorOwnedCertRequests)).
		Watches(&corev1.Service{}, handler.EnqueueRequestsFromMapFunc(r.NetworkPoliciesFromService)).
		WithEventFilter(predicate.GenerationChangedPredicate{}).
		Complete(r)
}

func (r *ApplicationReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	application := &skiperatorv1alpha1.Application{}
	err := r.GetClient().Get(ctx, req.NamespacedName, application)

	if errors.IsNotFound(err) {
		return reconcile.Result{}, nil
	} else if err != nil {
		r.GetRecorder().Eventf(
			application,
			corev1.EventTypeNormal, "ReconcileStartFail",
			"Something went wrong fetching the application. It might have been deleted",
		)
		return reconcile.Result{}, err
	}

	r.GetRecorder().Eventf(
		application,
		corev1.EventTypeNormal, "ReconcileStart",
		"Application "+application.Name+" has started reconciliation loop",
	)

	isApplicationMarkedToBeDeleted := application.GetDeletionTimestamp() != nil
	if isApplicationMarkedToBeDeleted {
		if ctrlutil.ContainsFinalizer(application, applicationFinalizer) {
			if err := r.finalizeApplication(ctx, application); err != nil {
				return ctrl.Result{}, err
			}

			ctrlutil.RemoveFinalizer(application, applicationFinalizer)
			err := r.GetClient().Update(ctx, application)
			if err != nil {
				return ctrl.Result{}, err
			}
		}
	}

	err = r.validateApplicationSpec(application)
	if err != nil {
		r.GetRecorder().Eventf(
			application,
			corev1.EventTypeWarning, "InvalidApplication",
			"Application failed validation and was rejected, error: %s", err.Error(),
		)
		return reconcile.Result{}, err
	}

	controllerDuties := []func(context.Context, *skiperatorv1alpha1.Application) (reconcile.Result, error){
		r.initializeApplicationStatus,
		r.initializeApplication,
		r.reconcileCertificate,
		r.reconcileDeployment,
		r.reconcileService,
		r.reconcileConfigMap,
		r.reconcileEgressServiceEntry,
		r.reconcileIngressGateway,
		r.reconcileIngressVirtualService,
		r.reconcileHorizontalPodAutoscaler,
		r.reconcilePeerAuthentication,
		r.reconcileServiceAccount,
		r.reconcileNetworkPolicy,
		r.reconcileAuthorizationPolicy,
		r.reconcilePodDisruptionBudget,
		r.reconcileServiceMonitor,
	}

	for _, fn := range controllerDuties {
		if _, err := fn(ctx, application); err != nil {
			return reconcile.Result{}, err
		}
	}

	r.GetRecorder().Eventf(
		application,
		corev1.EventTypeNormal, "ReconcileEnd",
		"Application "+application.Name+" has finished reconciliation loop",
	)

	return reconcile.Result{}, err
}

func (r *ApplicationReconciler) initializeApplication(ctx context.Context, application *skiperatorv1alpha1.Application) (reconcile.Result, error) {
	_ = r.GetClient().Get(ctx, types.NamespacedName{Namespace: application.Namespace, Name: application.Name}, application)

	if !ctrlutil.ContainsFinalizer(application, applicationFinalizer) {
		ctrlutil.AddFinalizer(application, applicationFinalizer)
	}

	application.Spec.Replicas.Max = util.Max(application.Spec.Replicas.Max, application.Spec.Replicas.Min)

	if len(application.Labels) == 0 {
		application.Labels = application.Spec.Labels
	} else {
		aggregateLabels := application.Labels
		maps.Copy(aggregateLabels, application.Spec.Labels)
		application.Labels = aggregateLabels
	}

	err := r.GetClient().Update(ctx, application)
	if err != nil {
		r.GetRecorder().Eventf(
			application,
			corev1.EventTypeNormal, "InitializeAppFunc",
			"Application "+application.Name+" could not init: "+err.Error(),
		)
	}

	return reconcile.Result{}, err
}

func (r *ApplicationReconciler) initializeApplicationStatus(ctx context.Context, application *skiperatorv1alpha1.Application) (reconcile.Result, error) {
	_ = r.GetClient().Get(ctx, types.NamespacedName{Namespace: application.Namespace, Name: application.Name}, application)

	application.FillDefaultsStatus()
	err := r.GetClient().Status().Update(ctx, application)
	if err != nil {
		r.GetRecorder().Eventf(
			application,
			corev1.EventTypeNormal, "InitializeAppStatusFunc",
			"Application "+application.Name+" could not init status: "+err.Error(),
		)
	}

	return reconcile.Result{}, err
}

func (r *ApplicationReconciler) finalizeApplication(ctx context.Context, application *skiperatorv1alpha1.Application) error {
	certificates, err := r.GetSkiperatorOwnedCertificates(ctx)
	if err != nil {
		return err
	}

	for _, certificate := range certificates.Items {
		if r.IsApplicationsCertificate(ctx, *application, certificate) {
			err = r.GetClient().Delete(ctx, &certificate)
			err = client.IgnoreNotFound(err)
			if err != nil {
				return err
			}
		}

	}
	return err
}

func (r *ApplicationReconciler) validateApplicationSpec(application *skiperatorv1alpha1.Application) error {
	validationFunctions := []func(application *skiperatorv1alpha1.Application) error{
		ValidateIngresses,
	}

	for _, validationFunction := range validationFunctions {
		err := validationFunction(application)

		if err != nil {
			return err
		}
	}

	return nil
}

// Name in the form of "servicemonitors.monitoring.coreos.com".
func (r *ApplicationReconciler) isCrdPresent(ctx context.Context, name string) bool {
	result, err := r.GetApiExtensionsClient().ApiextensionsV1().CustomResourceDefinitions().Get(ctx, name, metav1.GetOptions{})
	if err != nil || result == nil {
		return false
	}

	return true
}

func ValidateIngresses(application *skiperatorv1alpha1.Application) error {
	matchExpression, _ := regexp.Compile(`^([a-z0-9]+(-[a-z0-9]+)*\.)+[a-z]{2,}$`)
	for _, ingress := range application.Spec.Ingresses {
		if !matchExpression.MatchString(ingress) {
			errMessage := fmt.Sprintf("ingress with value '%s' was not valid. ingress must be lower case, contain no spaces, be a non-empty string, and have a hostname/domain separated by a period", ingress)
			return errors.NewInvalid(application.GroupVersionKind().GroupKind(), application.Name, field.ErrorList{
				field.Invalid(field.NewPath("application").Child("spec").Child("ingresses"), application.Spec.Ingresses, errMessage),
			})
		}
	}

	return nil
}
