package applicationcontroller

import (
	"context"

	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/util"
	securityv1beta1api "istio.io/api/security/v1beta1"
	typev1beta1 "istio.io/api/type/v1beta1"
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *ApplicationReconciler) reconcileAuthorizationPolicy(ctx context.Context, application *skiperatorv1alpha1.Application) (reconcile.Result, error) {
	controllerName := "AuthorizationPolicy"
	r.SetControllerProgressing(ctx, application, controllerName)

	authorizationPolicy := securityv1beta1.AuthorizationPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: application.Namespace,
			Name:      application.Name + "-actuator-deny",
		},
	}

	_, err := ctrlutil.CreateOrPatch(ctx, r.GetClient(), &authorizationPolicy, func() error {
		// Set application as owner of the authorizationPolicy
		err := ctrlutil.SetControllerReference(application, &authorizationPolicy, r.GetScheme())
		if err != nil {
			r.SetControllerError(ctx, application, controllerName, err)
			return err
		}

		r.SetLabelsFromApplication(ctx, &authorizationPolicy, *application)
		util.SetCommonAnnotations(&authorizationPolicy)

		// update authorizationPolicy rules and action
		authorizationPolicy.Spec = securityv1beta1api.AuthorizationPolicy{
			Action: securityv1beta1api.AuthorizationPolicy_DENY,
			Rules: []*securityv1beta1api.Rule{
				{
					To: []*securityv1beta1api.Rule_To{
						{
							Operation: &securityv1beta1api.Operation{
								Paths: []string{"/actuator*"},
							},
						},
					},
					From: []*securityv1beta1api.Rule_From{
						{
							Source: &securityv1beta1api.Source{
								Namespaces: []string{"istio-gateways"},
							},
						},
					},
				},
			},
			Selector: &typev1beta1.WorkloadSelector{
				MatchLabels: map[string]string{"app": application.Name},
			},
		}

		return nil
	})

	r.SetControllerFinishedOutcome(ctx, application, controllerName, err)

	return reconcile.Result{}, err
}
