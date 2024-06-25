package applicationcontroller

import (
	"context"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/gcp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/util"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type Config struct {
	Type                           string           `json:"type"`
	Audience                       string           `json:"audience"`
	ServiceAccountImpersonationUrl string           `json:"service_account_impersonation_url"`
	SubjectTokenType               string           `json:"subject_token_type"`
	TokenUrl                       string           `json:"token_url"`
	CredentialSource               CredentialSource `json:"credential_source"`
}
type CredentialSource struct {
	File string `json:"file"`
}

var controllerName = "ConfigMap"

func (r *ApplicationReconciler) reconcileConfigMap(ctx context.Context, application *skiperatorv1alpha1.Application) (reconcile.Result, error) {
	r.SetControllerProgressing(ctx, application, controllerName)

	if util.IsGCPAuthEnabled(application.Spec.GCP) {
		gcpIdentityConfigMapNamespacedName := types.NamespacedName{Namespace: "skiperator-system", Name: "gcp-identity-config"}
		gcpIdentityConfigMap, err := util.GetConfigMap(r.GetClient(), ctx, gcpIdentityConfigMapNamespacedName)

		if !util.ErrIsMissingOrNil(
			r.GetRecorder(),
			err,
			"Cannot find configmap named "+gcpIdentityConfigMapNamespacedName.Name+" in namespace "+gcpIdentityConfigMapNamespacedName.Namespace,
			application,
		) {
			r.SetControllerError(ctx, application, controllerName, err)
			return util.RequeueWithError(err)
		}

		err = r.setupGCPAuthConfigMap(ctx, gcpIdentityConfigMap, application)
		if err != nil {
			r.SetControllerError(ctx, application, controllerName, err)
			return util.RequeueWithError(err)
		}
	} else {
		gcpAuthConfigMap := corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: application.Namespace,
				Name:      gcp.GetGCPConfigMapName(application.Name),
			},
		}
		err := client.IgnoreNotFound(r.GetClient().Delete(ctx, &gcpAuthConfigMap))
		if err != nil {
			return util.RequeueWithError(err)
		}

	}

	r.SetControllerFinishedOutcome(ctx, application, controllerName, nil)

	return util.DoNotRequeue()

}

func (r *ApplicationReconciler) setupGCPAuthConfigMap(ctx context.Context, gcpIdentityConfigMap corev1.ConfigMap, application *skiperatorv1alpha1.Application) error {

	gcpAuthConfigMapName := gcp.GetGCPConfigMapName(application.Name)
	gcpAuthConfigMap, err := gcp.GetGoogleServiceAccountCredentialsConfigMap(
		ctx,
		application.Namespace,
		gcpAuthConfigMapName,
		application.Spec.GCP.Auth.ServiceAccount,
		gcpIdentityConfigMap,
	)
	if err != nil {
		return err
	}

	credentialsBytes := gcpAuthConfigMap.Data["config"]

	shouldReconcile, err := r.ShouldReconcile(ctx, &gcpAuthConfigMap)
	if err != nil || !shouldReconcile {
		r.SetControllerFinishedOutcome(ctx, application, controllerName, err)
		return err
	}

	_, err = ctrlutil.CreateOrPatch(ctx, r.GetClient(), &gcpAuthConfigMap, func() error {
		// Set application as owner of the configmap
		err := ctrlutil.SetControllerReference(application, &gcpAuthConfigMap, r.GetScheme())
		if err != nil {
			r.SetControllerError(ctx, application, controllerName, err)
			return err
		}
		r.SetLabelsFromApplication(&gcpAuthConfigMap, *application)
		gcpAuthConfigMap.Data["config"] = credentialsBytes
		return nil
	})

	return err
}
