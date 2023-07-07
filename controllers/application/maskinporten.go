package applicationcontroller

import (
	"context"

	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/util"
	nais_io_v1 "github.com/nais/liberator/pkg/apis/nais.io/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *ApplicationReconciler) reconcileMaskinporten(ctx context.Context, application *skiperatorv1alpha1.Application) (reconcile.Result, error) {
	controllerName := "Maskinporten"
	r.SetControllerProgressing(ctx, application, controllerName)

	secretName, err := util.GetSecretName("maskinporten", application.Name)
	if err != nil {
		return reconcile.Result{}, err
	}

	maskinporten := nais_io_v1.MaskinportenClient{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "nais.io/v1",
			Kind:       "MaskinportenClient",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: application.Namespace,
			Name:      application.Name,
		},
	}

	if maskinportenSpecifiedInSpec(application.Spec.Maskinporten) {
		_, err = ctrlutil.CreateOrPatch(ctx, r.GetClient(), &maskinporten, func() error {
			err := ctrlutil.SetControllerReference(application, &maskinporten, r.GetScheme())
			if err != nil {
				r.SetControllerError(ctx, application, controllerName, err)
				return err
			}

			r.SetLabelsFromApplication(ctx, &maskinporten, *application)
			util.SetCommonAnnotations(&maskinporten)

			scopes := nais_io_v1.MaskinportenScope{}
			if application.Spec.Maskinporten.Scopes != nil {
				scopes = *application.Spec.Maskinporten.Scopes
			}

			maskinporten.Spec = nais_io_v1.MaskinportenClientSpec{
				SecretName: secretName,
				Scopes:     scopes,
			}

			return nil
		})

		if err != nil {
			r.SetControllerError(ctx, application, controllerName, err)
			return reconcile.Result{}, err
		}
	} else {
		err = r.GetClient().Delete(ctx, &maskinporten)
		err = client.IgnoreNotFound(err)
		if err != nil {
			r.SetControllerError(ctx, application, controllerName, err)
			return reconcile.Result{}, err
		}
	}

	r.SetControllerFinishedOutcome(ctx, application, controllerName, err)

	return reconcile.Result{}, err
}

func maskinportenSpecifiedInSpec(mp *skiperatorv1alpha1.Maskinporten) bool {
	return mp != nil && mp.Enabled
}
