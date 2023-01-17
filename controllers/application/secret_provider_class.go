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
	secretsStore "sigs.k8s.io/secrets-store-csi-driver/apis/v1alpha1"
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

		secretProviderClass := secretsStore.SecretProviderClass{ObjectMeta: metav1.ObjectMeta{Namespace: application.Namespace, Name: name}}
		_, err = ctrlutil.CreateOrPatch(ctx, r.GetClient(), &secretProviderClass, func() error {
			secretProviderClass.Spec.Parameters = map[string]string{
				"secrets": fmt.Sprintf("|\n  - resourceName: \"%s\"\n    path: \"%s.txt\"", secretManagerReference, secretName),
			}
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
