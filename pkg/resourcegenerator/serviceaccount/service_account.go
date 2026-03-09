package serviceaccount

import (
	"maps"

	"github.com/kartverket/skiperator/pkg/reconciliation"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/resourceutils/generator"
	corev1 "k8s.io/api/core/v1"
)

var multiGenerator = generator.NewMulti()

func Generate(r reconciliation.Reconciliation) error {
	return multiGenerator.Generate(r, "ServiceAccount")
}

func setGCPSAAnnotation(serviceAccount *corev1.ServiceAccount, saEmail string) {
	annotations := serviceAccount.GetAnnotations()
	if len(annotations) == 0 {
		annotations = make(map[string]string)
	}
	maps.Copy(annotations, map[string]string{
		"iam.gke.io/gcp-service-account": saEmail,
	})
	serviceAccount.SetAnnotations(annotations)
}
