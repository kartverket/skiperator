package applicationcontroller

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	policyv1 "k8s.io/api/policy/v1"

	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
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
		Owns(&networkingv1beta1.ServiceEntry{}).
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
		Owns(&pov1.ServiceMonitor{}).
		Watches(&certmanagerv1.Certificate{}, handler.EnqueueRequestsFromMapFunc(r.SkiperatorOwnedCertRequests)).
		WithEventFilter(predicate.Or(predicate.GenerationChangedPredicate{}, predicate.LabelChangedPredicate{})).
		Complete(r)
}

func (r *ApplicationReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	application := &skiperatorv1alpha1.Application{}
	err := r.GetClient().Get(ctx, req.NamespacedName, application)

	if errors.IsNotFound(err) {
		return reconcile.Result{}, nil
	} else if err != nil {
		r.EmitWarningEvent(application, "ReconcileStartFail", "something went wrong fetching the application, it might have been deleted")
		return reconcile.Result{}, err
	}

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
		r.EmitNormalEvent(application, "InvalidApplication", fmt.Sprintf("Application %v failed validation and was rejected, error: %s", application.Name, err.Error()))
		return reconcile.Result{}, err
	}

	tmpApplication := application.DeepCopy()
	application.FillDefaultsSpec()
	if !ctrlutil.ContainsFinalizer(application, applicationFinalizer) {
		ctrlutil.AddFinalizer(application, applicationFinalizer)
	}

	if len(application.Labels) == 0 {
		application.Labels = application.Spec.Labels
	} else {
		aggregateLabels := application.Labels
		maps.Copy(aggregateLabels, application.Spec.Labels)
		application.Labels = aggregateLabels
	}

	application.FillDefaultsStatus()

	specDiff, err := util.GetObjectDiff(tmpApplication.Spec, application.Spec)
	if err != nil {
		return reconcile.Result{}, err
	}

	statusDiff, err := util.GetObjectDiff(tmpApplication.Status, application.Status)
	if err != nil {
		return reconcile.Result{}, err
	}

	// If we update the Application initially on applied defaults before starting reconciling resources we allow all
	// updates to be visible even though the controllerDuties may take some time.
	if len(statusDiff) > 0 {
		err := r.GetClient().Status().Update(ctx, application)
		return reconcile.Result{Requeue: true}, err
	}

	// Finalizer check is due to a bug when updating using controller-runtime
	// See https://github.com/kubernetes-sigs/controller-runtime/issues/2453
	if len(specDiff) > 0 || (!ctrlutil.ContainsFinalizer(tmpApplication, applicationFinalizer) && ctrlutil.ContainsFinalizer(application, applicationFinalizer)) {
		err := r.GetClient().Update(ctx, application)
		return reconcile.Result{Requeue: true}, err
	}

	r.EmitNormalEvent(application, "ReconcileStart", fmt.Sprintf("Application %v has started reconciliation loop", application.Name))

	controllerDuties := []func(context.Context, *skiperatorv1alpha1.Application) (reconcile.Result, error){
		r.reconcileCertificate,
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
		r.reconcileDeployment,
	}

	for _, fn := range controllerDuties {
		res, err := fn(ctx, application)
		if err != nil {
			return res, err
		} else if res.RequeueAfter > 0 || res.Requeue {
			return res, nil
		}
	}
	//r.GetClient().Status().Update(ctx, application)
	r.EmitNormalEvent(application, "ReconcileEnd", fmt.Sprintf("Application %v has finished reconciliation loop", application.Name))

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

func (r *ApplicationReconciler) manageControllerStatus(context context.Context, app *skiperatorv1alpha1.Application, controller string, statusName skiperatorv1alpha1.StatusNames, message string) (reconcile.Result, error) {
	app.UpdateControllerStatus(controller, message, statusName)
	err := r.GetClient().Status().Update(context, app)
	if err != nil {
		return reconcile.Result{Requeue: true}, err
	}
	return reconcile.Result{Requeue: true}, nil
}

func (r *ApplicationReconciler) manageControllerStatusError(context context.Context, app *skiperatorv1alpha1.Application, controller string, issue error) (reconcile.Result, error) {
	app.UpdateControllerStatus(controller, issue.Error(), skiperatorv1alpha1.ERROR)
	err := r.GetClient().Status().Update(context, app)
	r.EmitWarningEvent(app, "ControllerFault", fmt.Sprintf("%v controller experienced an error: %v", controller, issue.Error()))

	if err != nil {
		return reconcile.Result{}, err
	}
	return reconcile.Result{}, issue
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

type ControllerResources string

const (
	DEPLOYMENT              ControllerResources = "Deployment"
	POD                     ControllerResources = "Pod"
	SERVICE                 ControllerResources = "Service"
	SERVICEACCOUNT          ControllerResources = "ServiceAccount"
	CONFIGMAP               ControllerResources = "ConfigMap"
	NETWORKPOLICY           ControllerResources = "NetworkPolicy"
	GATEWAY                 ControllerResources = "Gateway"
	SERVICEENTRY            ControllerResources = "ServiceEntry"
	VIRTUALSERVICE          ControllerResources = "VirtualService"
	PEERAUTHENTICATION      ControllerResources = "PeerAuthentication"
	HORIZONTALPODAUTOSCALER ControllerResources = "HorizontalPodAutoscaler"
	CERTIFICATE             ControllerResources = "Certificate"
	AUTHORIZATIONPOLICY     ControllerResources = "AuthorizationPolicy"
)

var GroupKindFromControllerResource = map[string]metav1.GroupKind{
	"deployment": {
		Group: "apps",
		Kind:  string(DEPLOYMENT),
	},
	"pod": {
		Group: "",
		Kind:  string(POD),
	},
	"service": {
		Group: "",
		Kind:  string(SERVICE),
	},
	"serviceaccount": {
		Group: "",
		Kind:  string(SERVICEACCOUNT),
	},
	"configmaps": {
		Group: "",
		Kind:  string(CONFIGMAP),
	},
	"networkpolicy": {
		Group: "networking.k8s.io",
		Kind:  string(NETWORKPOLICY),
	},
	"gateway": {
		Group: "networking.istio.io",
		Kind:  string(GATEWAY),
	},
	"serviceentry": {
		Group: "networking.istio.io",
		Kind:  string(SERVICEENTRY),
	},
	"virtualservice": {
		Group: "networking.istio.io",
		Kind:  string(VIRTUALSERVICE),
	},
	"peerauthentication": {
		Group: "security.istio.io",
		Kind:  string(PEERAUTHENTICATION),
	},
	"horizontalpodautoscaler": {
		Group: "autoscaling",
		Kind:  string(HORIZONTALPODAUTOSCALER),
	},
	"certificate": {
		Group: "cert-manager.io",
		Kind:  string(CERTIFICATE),
	},
	"authorizationpolicy": {
		Group: "security.istio.io",
		Kind:  string(AUTHORIZATIONPOLICY),
	},
}

func (r *ApplicationReconciler) setResourceLabelsIfApplies(obj client.Object, app skiperatorv1alpha1.Application) {
	objectGroupVersionKind := obj.GetObjectKind().GroupVersionKind()

	for controllerResource, resourceLabels := range app.Spec.ResourceLabels {
		resourceLabelGroupKind, present := GroupKindFromControllerResource[strings.ToLower(controllerResource)]
		if present {
			if strings.EqualFold(objectGroupVersionKind.Group, resourceLabelGroupKind.Group) && strings.EqualFold(objectGroupVersionKind.Kind, resourceLabelGroupKind.Kind) {
				objectLabels := obj.GetLabels()
				if len(objectLabels) == 0 {
					objectLabels = make(map[string]string)
				}
				maps.Copy(objectLabels, resourceLabels)
				obj.SetLabels(objectLabels)
			}
		} else {
			r.EmitWarningEvent(&app, "MistypedLabel", fmt.Sprintf("could not find according Kind for Resource %v, make sure your resource is spelled correctly", controllerResource))
		}
	}
}

func (r *ApplicationReconciler) SetLabelsFromApplication(object client.Object, app skiperatorv1alpha1.Application) {
	labels := object.GetLabels()
	if len(labels) == 0 {
		labels = make(map[string]string)
	}
	if app.Spec.Labels != nil {
		maps.Copy(labels, app.Spec.Labels)
		object.SetLabels(labels)
	}

	r.setResourceLabelsIfApplies(object, app)
}
