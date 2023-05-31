package applicationcontroller

import (
	"context"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/util"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *ApplicationReconciler) reconcileServiceAccount(ctx context.Context, application *skiperatorv1alpha1.Application) (reconcile.Result, error) {
	controllerName := "ServiceAccount"
	r.SetControllerProgressing(ctx, application, controllerName)

	serviceAccount := corev1.ServiceAccount{ObjectMeta: metav1.ObjectMeta{Namespace: application.Namespace, Name: application.Name}}
	r.SetLabelsFromApplication(ctx, &serviceAccount, *application)

	err := CreateOrPatchServiceAccount(ctx, serviceAccount, application, r.GetClient(), r.GetScheme())

	r.SetControllerFinishedOutcome(ctx, application, controllerName, err)

	return reconcile.Result{}, err
}

func CreateOrPatchServiceAccount(context context.Context, serviceAccount corev1.ServiceAccount, ownerObject metav1.Object, client client.Client, scheme *runtime.Scheme) error {
	_, err := ctrlutil.CreateOrPatch(context, client, &serviceAccount, func() error {
		// Set application as owner of the sidecar
		err := ctrlutil.SetControllerReference(ownerObject, &serviceAccount, scheme)
		if err != nil {
			return err
		}

		util.SetCommonAnnotations(&serviceAccount)

		return nil
	})

	return err
}
