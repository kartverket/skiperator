package applicationcontroller

import (
	"context"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/istio"

	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/util"
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *ApplicationReconciler) reconcilePeerAuthentication(ctx context.Context, application *skiperatorv1alpha1.Application) (reconcile.Result, error) {
	controllerName := "PeerAuthentication"
	r.SetControllerProgressing(ctx, application, controllerName)

	peerAuthentication := securityv1beta1.PeerAuthentication{ObjectMeta: metav1.ObjectMeta{Namespace: application.Namespace, Name: application.Name}}
	shouldReconcile, err := r.ShouldReconcile(ctx, &peerAuthentication)
	if err != nil {
		r.SetControllerFinishedOutcome(ctx, application, controllerName, err)
		return reconcile.Result{}, err
	}

	if !shouldReconcile {
		r.SetControllerFinishedOutcome(ctx, application, controllerName, err)
		return reconcile.Result{}, nil
	}

	_, err = ctrlutil.CreateOrPatch(ctx, r.GetClient(), &peerAuthentication, func() error {
		// Set application as owner of the peer authentication
		err := ctrlutil.SetControllerReference(application, &peerAuthentication, r.GetScheme())
		if err != nil {
			r.SetControllerError(ctx, application, controllerName, err)
			return err
		}

		peerAuthentication.Spec = istio.GetPeerAuthentication(application.Name)

		r.SetLabelsFromApplication(&peerAuthentication, *application)
		util.SetCommonAnnotations(&peerAuthentication)

		return nil
	})

	r.SetControllerFinishedOutcome(ctx, application, controllerName, err)

	return reconcile.Result{}, err
}
