package testutil

import (
	"context"

	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/internal/config"
	"github.com/kartverket/skiperator/pkg/log"
	"github.com/kartverket/skiperator/pkg/reconciliation"
	"golang.org/x/exp/maps"
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
	maps.Copy(application.Labels, application.GetDefaultLabels())
	ctx := context.TODO()
	r := reconciliation.NewApplicationReconciliation(ctx, application, log.NewLogger(), false, nil, nil, config.SkiperatorConfig{})

	return r
}
