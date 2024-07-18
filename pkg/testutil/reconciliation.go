package testutil

import (
	"context"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/log"
	"github.com/kartverket/skiperator/pkg/reconciliation"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/resourceutils"
	"golang.org/x/exp/maps"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetTestMinimalAppReconciliation() *reconciliation.ApplicationReconciliation {
	application := &skiperatorv1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "minimal",
			Namespace: "test",
			Labels:    make(map[string]string),
		},
	}
	application.Spec = skiperatorv1alpha1.ApplicationSpec{
		Image: "image",
		Port:  8080,
	}
	application.FillDefaultsSpec()
	maps.Copy(application.Labels, resourceutils.GetApplicationDefaultLabels(application))
	identityConfigMap := corev1.ConfigMap{}
	identityConfigMap.Data = map[string]string{"workloadIdentityPool": "test-pool"}
	ctx := context.TODO()
	r := reconciliation.NewApplicationReconciliation(ctx, application, log.FromContext(ctx), nil, &identityConfigMap)

	return r
}
