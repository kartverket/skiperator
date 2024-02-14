package applicationcontroller

import (
	"context"
	"github.com/kartverket/skiperator/api/v1alpha1/digdirator"

	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/util"
	naisiov1 "github.com/nais/liberator/pkg/apis/nais.io/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *ApplicationReconciler) reconcileMaskinporten(ctx context.Context, application *skiperatorv1alpha1.Application) (reconcile.Result, error) {
	controllerName := "Maskinporten"
	r.SetControllerProgressing(ctx, application, controllerName)

	var err error

	maskinporten := naisiov1.MaskinportenClient{
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
				return err
			}

			r.SetLabelsFromApplication(&maskinporten, *application)
			util.SetCommonAnnotations(&maskinporten)

			maskinporten.Spec, err = getMaskinportenSpec(application)
			return err
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

func getMaskinportenSpec(application *skiperatorv1alpha1.Application) (naisiov1.MaskinportenClientSpec, error) {
	secretName, err := getMaskinportenSecretName(application.Name)
	if err != nil {
		return naisiov1.MaskinportenClientSpec{}, err
	}

	scopes := naisiov1.MaskinportenScope{}
	if application.Spec.Maskinporten.Scopes != nil {
		scopes = *application.Spec.Maskinporten.Scopes
	}

	return naisiov1.MaskinportenClientSpec{
		ClientName: getClientNameMaskinporten(application.Name, application.Spec.Maskinporten),
		SecretName: secretName,
		Scopes:     scopes,
	}, nil
}

func getClientNameMaskinporten(applicationName string, maskinportenSettings *digdirator.Maskinporten) string {
	if maskinportenSettings.ClientName != nil {
		return *maskinportenSettings.ClientName
	}

	return applicationName
}

func maskinportenSpecifiedInSpec(maskinportenSettings *digdirator.Maskinporten) bool {
	return maskinportenSettings != nil && maskinportenSettings.Enabled
}

func getMaskinportenSecretName(name string) (string, error) {
	return util.GetSecretName("maskinporten", name)
}
