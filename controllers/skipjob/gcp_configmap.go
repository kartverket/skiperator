package skipjobcontroller

import (
	"context"
	"fmt"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/gcp"
	"github.com/kartverket/skiperator/pkg/util"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *SKIPJobReconciler) reconcileConfigMap(ctx context.Context, skipJob *skiperatorv1alpha1.SKIPJob) (reconcile.Result, error) {
	if skipJob.Spec.Container.GCP != nil {
		gcpIdentityConfigMapNamespacedName := types.NamespacedName{Namespace: "skiperator-system", Name: "gcp-identity-config"}
		gcpIdentityConfigMap, err := util.GetConfigMap(r.GetClient(), ctx, gcpIdentityConfigMapNamespacedName)

		if !util.ErrIsMissingOrNil(
			r.GetRecorder(),
			err,
			fmt.Sprintf("cannot find configmap named %v in namespace %v", gcpIdentityConfigMapNamespacedName.Name, gcpIdentityConfigMapNamespacedName.Namespace),
			skipJob,
		) {
			return util.RequeueWithError(err)
		}

		err = r.setupGCPAuthConfigMap(ctx, gcpIdentityConfigMap, skipJob)
		if err != nil {
			return util.RequeueWithError(err)
		}
	} else {
		gcpAuthConfigMap := corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: skipJob.Namespace,
				Name:      gcp.GetGCPConfigMapName(skipJob.Name),
			},
		}
		err := r.DeleteObjectIfExists(ctx, &gcpAuthConfigMap)
		if err != nil {
			return util.RequeueWithError(err)
		}

	}
	return util.DoNotRequeue()

}

func (r *SKIPJobReconciler) setupGCPAuthConfigMap(ctx context.Context, gcpIdentityConfigMap corev1.ConfigMap, skipJob *skiperatorv1alpha1.SKIPJob) error {

	gcpAuthConfigMapName := gcp.GetGCPConfigMapName(skipJob.Name)
	gcpAuthConfigMap, err := gcp.GetGoogleServiceAccountCredentialsConfigMap(
		ctx,
		skipJob.Namespace,
		gcpAuthConfigMapName,
		skipJob.Spec.Container.GCP.Auth.ServiceAccount,
		gcpIdentityConfigMap,
	)
	if err != nil {
		return err
	}

	_, err = ctrlutil.CreateOrPatch(ctx, r.GetClient(), &gcpAuthConfigMap, func() error {
		// Set application as owner of the configmap
		err := ctrlutil.SetControllerReference(skipJob, &gcpAuthConfigMap, r.GetScheme())
		if err != nil {
			return err
		}

		return nil
	})

	return err
}
