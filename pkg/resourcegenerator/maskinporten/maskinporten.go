package maskinporten

import (
	"fmt"

	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/api/v1alpha1/digdirator"
	"github.com/kartverket/skiperator/pkg/reconciliation"
	"github.com/kartverket/skiperator/pkg/util"
	naisiov1 "github.com/nais/liberator/pkg/apis/nais.io/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Generate(r reconciliation.Reconciliation) error {
	ctxLog := r.GetLogger()
	if r.GetType() != reconciliation.ApplicationType {
		err := &util.SubResourceError{Message: "Unsupported type in maskin porten resource", WrapErr: fmt.Errorf("unsupported type %s in maskin porten resource", r.GetType()), Reason: util.UnsupportedTypeResource}
		return err
	}
	application, ok := r.GetSKIPObject().(*skiperatorv1alpha1.Application)
	if !ok {
		err := &util.SubResourceError{Message: "Failed to generate maskin porten resource", WrapErr: fmt.Errorf("failed to cast resource to application"), Reason: util.InternalError}
		ctxLog.Error(err, err.Message)
		return err
	}

	if !MaskinportenSpecifiedInSpec(application.Spec.Maskinporten) {
		ctxLog.Debug("Maskinporten not specified in spec, skipping generation")
		return nil
	}

	ctxLog.Debug("Attempting to generate maskin porten resource  for application", "application", application.Name)

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

	maskinporten.Spec, err = getMaskinportenSpec(application)
	if err != nil {
		err := &util.SubResourceError{Message: "Failed to generate maskin porten resource spec", WrapErr: err, Reason: util.ResourceDependencyNotFound}
		return err
	}

	r.AddResource(&maskinporten)
	ctxLog.Debug("Finished generating maskin porten resource for application", "application", application.Name)
	return nil
}

func getMaskinportenSpec(application *skiperatorv1alpha1.Application) (naisiov1.MaskinportenClientSpec, error) {
	secretName, err := GetMaskinportenSecretName(application.Name)
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

func MaskinportenSpecifiedInSpec(maskinportenSettings *digdirator.Maskinporten) bool {
	return maskinportenSettings != nil && maskinportenSettings.Enabled
}

func GetMaskinportenSecretName(name string) (string, error) {
	return util.GetSecretName("maskinporten", name)
}
