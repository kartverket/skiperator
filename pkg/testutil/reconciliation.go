package testutil

import (
	"context"

	skiperatorv1beta1 "github.com/kartverket/skiperator/api/v1beta1"
	"github.com/kartverket/skiperator/pkg/log"
	"github.com/kartverket/skiperator/pkg/reconciliation"
	"golang.org/x/exp/maps"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetTestMinimalAppReconciliation() *reconciliation.ApplicationReconciliation {
	application := &skiperatorv1beta1.Application{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "minimal",
			Namespace: "test",
			Labels:    make(map[string]string),
		},
	}
	application.Spec = skiperatorv1beta1.ApplicationSpec{
		Image: "image",
		Port:  8080,
	}
	application.FillDefaultsSpec()
	maps.Copy(application.Labels, application.GetDefaultLabels())
	identityConfigMap := corev1.ConfigMap{}
	identityConfigMap.Data = map[string]string{"workloadIdentityPool": "test-pool"}
	ctx := context.TODO()
	r := reconciliation.NewApplicationReconciliation(ctx, application, log.NewLogger(), false, nil, &identityConfigMap, nil)

	return r
}
