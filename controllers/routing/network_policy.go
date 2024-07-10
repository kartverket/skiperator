package routingcontroller

import (
	"context"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/reconciliation"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func Generate(r reconciliation.Reconciliation) error {

	// Get map of unique network policies: map[networkPolicyName]targetApp

	ctxLog.Debug("Finished creating certificates for application", application.Name)
	return nil
}

func getApplication(client client.Client, ctx context.Context, namespacedName types.NamespacedName) (skiperatorv1alpha1.Application, error) {
	application := skiperatorv1alpha1.Application{}

	err := client.Get(ctx, namespacedName, &application)

	return application, err
}
