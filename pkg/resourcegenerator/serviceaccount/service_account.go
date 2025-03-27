package serviceaccount

import (
	"maps"

	skiperatorv1alpha1 "github.com/kartverket/skiperator/v3/api/v1alpha1"
	"github.com/kartverket/skiperator/v3/pkg/reconciliation"
	"github.com/kartverket/skiperator/v3/pkg/resourcegenerator/resourceutils/generator"
	corev1 "k8s.io/api/core/v1"
)

var multiGenerator = generator.NewMulti()

func Generate(r reconciliation.Reconciliation) error {
	return multiGenerator.Generate(r, "ServiceAccount")
}

func setCloudSqlAnnotations(serviceAccount *corev1.ServiceAccount, gcp skiperatorv1alpha1.SKIPObject) {
	annotations := serviceAccount.GetAnnotations()
	if len(annotations) == 0 {
		annotations = make(map[string]string)
	}
	maps.Copy(annotations, map[string]string{
		"iam.gke.io/gcp-service-account": gcp.GetCommonSpec().GCP.CloudSQLProxy.ServiceAccount,
	})
	serviceAccount.SetAnnotations(annotations)
}
