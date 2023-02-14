package applicationcontroller

import (
	"context"

	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	securityv1beta1api "istio.io/api/security/v1beta1"
	typev1beta1 "istio.io/api/type/v1beta1"
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *ApplicationReconciler) reconcilePeerAuthentication(ctx context.Context, application *skiperatorv1alpha1.Application) (reconcile.Result, error) {
	controllerName := "PeerAuthentication"
	r.SetControllerProgressing(ctx, application, controllerName)

	peerAuthentication := securityv1beta1.PeerAuthentication{ObjectMeta: metav1.ObjectMeta{Namespace: application.Namespace, Name: application.Name}}
	_, err := ctrlutil.CreateOrPatch(ctx, r.GetClient(), &peerAuthentication, func() error {
		// Set application as owner of the peer authentication
		err := ctrlutil.SetControllerReference(application, &peerAuthentication, r.GetScheme())
		if err != nil {
			r.SetControllerError(ctx, application, controllerName, err)
			return err
		}

		r.SetLabelsFromApplication(ctx, &peerAuthentication, *application)
		peerAuthentication.ObjectMeta.Annotations = map[string]string{
			"argocd.argoproj.io/sync-options": "Prune=false",
		}

		peerAuthentication.Spec.Selector = &typev1beta1.WorkloadSelector{}
		labels := map[string]string{"app": application.Name}
		peerAuthentication.Spec.Selector.MatchLabels = labels

		peerAuthentication.Spec.Mtls = &securityv1beta1api.PeerAuthentication_MutualTLS{}
		peerAuthentication.Spec.Mtls.Mode = securityv1beta1api.PeerAuthentication_MutualTLS_STRICT

		return nil
	})

	r.SetControllerFinishedOutcome(ctx, application, controllerName, err)

	return reconcile.Result{}, err
}
