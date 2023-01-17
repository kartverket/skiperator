package applicationcontroller

import (
	"context"
	"fmt"
	"hash/fnv"
	"strings"

	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	secretsStorev1 "sigs.k8s.io/secrets-store-csi-driver/apis/v1"
)

func (r *ApplicationReconciler) reconcileSecretProviderClass(ctx context.Context, application *skiperatorv1alpha1.Application) (reconcile.Result, error) {
	controllerName := "SecretProviderClass"
	r.SetControllerProgressing(ctx, application, controllerName)

	var err error

	var secretManagerReferences []string
	for _, file := range application.Spec.FilesFrom {
		value := file.GcpSecretManager
		if len(value) > 0 && !contains(secretManagerReferences, value) {
			secretManagerReferences = append(secretManagerReferences, value)
		}
	}
	for _, env := range application.Spec.EnvFrom {
		value := env.GcpSecretManager
		if len(value) > 0 && !contains(secretManagerReferences, value) {
			secretManagerReferences = append(secretManagerReferences, value)
		}
	}

	for _, secretManagerReference := range secretManagerReferences {
		hash := fnv.New64()
		_, _ = hash.Write([]byte(secretManagerReference))
		name := fmt.Sprintf("%s-%x", application.Name, hash.Sum64())

		// projects/$PROJECT_ID/secrets/testsecret/versions/latest
		secretName := strings.Split(secretManagerReference, "/")[3]

		secretProviderClass := secretsStorev1.SecretProviderClass{ObjectMeta: metav1.ObjectMeta{Namespace: application.Namespace, Name: name}}
		_, err = ctrlutil.CreateOrPatch(ctx, r.GetClient(), &secretProviderClass, func() error {
			err := ctrlutil.SetControllerReference(application, &secretProviderClass, r.GetScheme())
			if err != nil {
				r.SetControllerError(ctx, application, controllerName, err)
				return err
			}

			r.SetLabelsFromApplication(ctx, &secretProviderClass, *application)

			secretProviderClass.Spec.Provider = "gcp"
			// Create mapping for secret in GCP to path in kubernetes
			secretProviderClass.Spec.Parameters = map[string]string{
				"secrets": fmt.Sprintf(
					"- resourceName: \"%s\"\n"+
						"  path: \"%s\"\n"+
						"  objectAlias: \"secretalias\"",
					secretManagerReference,
					secretName,
				),
			}
			// Create kubernetes secret as well
			secretProviderClass.Spec.SecretObjects = []*secretsStorev1.SecretObject{{
				SecretName: fmt.Sprintf("gcp-sm-%s", secretName),
				Type:       "Opaque",
				Labels: map[string]string{
					"app.kubernetes.io/managed-by": "skiperator",
				},
				Data: []*secretsStorev1.SecretObjectData{{
					Key:        secretName,
					ObjectName: "secretalias",
				}},
			}}
			return nil
		})
	}

	r.SetControllerFinishedOutcome(ctx, application, controllerName, err)

	return reconcile.Result{}, err
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
