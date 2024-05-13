package applicationcontroller

import (
	"context"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/util"
	"golang.org/x/exp/maps"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *ApplicationReconciler) reconcileServiceAccount(ctx context.Context, application *skiperatorv1alpha1.Application) (reconcile.Result, error) {
	controllerName := "ServiceAccount"
	r.SetControllerProgressing(ctx, application, controllerName)

	serviceAccount := corev1.ServiceAccount{ObjectMeta: metav1.ObjectMeta{Namespace: application.Namespace, Name: application.Name}}

	shouldReconcile, err := r.ShouldReconcile(ctx, &serviceAccount)
	if err != nil || !shouldReconcile {
		r.SetControllerFinishedOutcome(ctx, application, controllerName, err)
		return util.RequeueWithError(err)
	}

	_, err = ctrlutil.CreateOrPatch(ctx, r.GetClient(), &serviceAccount, func() error {
		// Set application as owner of the sidecar
		err := ctrlutil.SetControllerReference(application, &serviceAccount, r.GetScheme())
		if err != nil {
			r.SetControllerError(ctx, application, controllerName, err)
			return err
		}

		if util.IsCloudSqlProxyEnabled(application.Spec.GCP) {
			setCloudSqlAnnotations(&serviceAccount, application)
		}

		r.SetLabelsFromApplication(&serviceAccount, *application)
		util.SetCommonAnnotations(&serviceAccount)

		return nil
	})

	r.SetControllerFinishedOutcome(ctx, application, controllerName, err)

	return util.RequeueWithError(err)
}

func setCloudSqlAnnotations(serviceAccount *corev1.ServiceAccount, application *skiperatorv1alpha1.Application) {
	annotations := serviceAccount.GetAnnotations()
	if len(annotations) == 0 {
		annotations = make(map[string]string)
	}
	maps.Copy(annotations, map[string]string{
		"iam.gke.io/gcp-service-account": application.Spec.GCP.CloudSQLProxy.ServiceAccount,
	})
	serviceAccount.SetAnnotations(annotations)
}
