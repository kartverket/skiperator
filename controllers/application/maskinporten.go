package applicationcontroller

import (
	"context"
	"fmt"

	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/util"
	nais_io_v1 "github.com/nais/liberator/pkg/apis/nais.io/v1"
	"github.com/nais/liberator/pkg/namegen"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *ApplicationReconciler) reconcileMaskinporten(ctx context.Context, application *skiperatorv1alpha1.Application) (reconcile.Result, error) {
	controllerName := "Maskinporten"
	r.SetControllerProgressing(ctx, application, controllerName)

	secretName, err := namegen.ShortName(fmt.Sprintf("maskinporten-%s/%s", application.Namespace, application.Name), validation.DNS1035LabelMaxLength)
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

			maskinporten.Spec = nais_io_v1.MaskinportenClientSpec{
				SecretName: secretName,
				Scopes:     application.Spec.Maskinporten.Scopes,
			}

			return nil
		})
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

func maskinportenSpecifiedInSpec(mp *nais_io_v1.Maskinporten) bool {
	return mp != nil && mp.Enabled
}
