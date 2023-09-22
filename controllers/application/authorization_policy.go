package applicationcontroller

import (
	"context"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/util"
	securityv1beta1api "istio.io/api/security/v1beta1"
	typev1beta1 "istio.io/api/type/v1beta1"
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *ApplicationReconciler) reconcileAuthorizationPolicy(ctx context.Context, application *skiperatorv1alpha1.Application) (reconcile.Result, error) {
	controllerName := "AuthorizationPolicy"
	r.SetControllerProgressing(ctx, application, controllerName)

	defaultDenyPaths := []string{
		"/actuator*",
	}
	defaultDenyAuthPolicy := getDefaultDenyPolicy(application, defaultDenyPaths)

	shouldReconcile, err := r.ShouldReconcile(ctx, &defaultDenyAuthPolicy)
	if err != nil {
		r.SetControllerFinishedOutcome(ctx, application, controllerName, err)
		return reconcile.Result{}, err
	}

	if !shouldReconcile {
		r.SetControllerFinishedOutcome(ctx, application, controllerName, err)
		return reconcile.Result{}, nil
	}

	if application.Spec.AuthorizationSettings != nil {
		if application.Spec.AuthorizationSettings.AllowAll == true {
			err := r.GetClient().Delete(ctx, &defaultDenyAuthPolicy)
			err = client.IgnoreNotFound(err)
			if err != nil {
				r.SetControllerError(ctx, application, controllerName, err)
				return reconcile.Result{}, err
			}

			r.SetControllerFinishedOutcome(ctx, application, controllerName, nil)
			return reconcile.Result{}, nil
		}
	}

	_, err = ctrlutil.CreateOrPatch(ctx, r.GetClient(), &defaultDenyAuthPolicy, func() error {
		err := ctrlutil.SetControllerReference(application, &defaultDenyAuthPolicy, r.GetScheme())
		if err != nil {
			r.SetControllerError(ctx, application, controllerName, err)
			return err
		}

		r.SetLabelsFromApplication(&defaultDenyAuthPolicy, *application)
		util.SetCommonAnnotations(&defaultDenyAuthPolicy)

		if application.Spec.AuthorizationSettings != nil {

			// As of now we only use one rule and one operation for all default denies. No need to loop over them all
			defaultDenyToOperation := defaultDenyAuthPolicy.Spec.Rules[0].To[0].Operation
			defaultDenyToOperation.NotPaths = nil

			if len(application.Spec.AuthorizationSettings.AllowList) > 0 {
				for _, endpoint := range application.Spec.AuthorizationSettings.AllowList {
					defaultDenyToOperation.NotPaths = append(defaultDenyToOperation.NotPaths, endpoint)
				}
			}
		}

		// update defaultDenyAuthPolicy rules and action
		return nil
	})

	r.SetControllerFinishedOutcome(ctx, application, controllerName, err)

	return reconcile.Result{}, err
}

func getGeneralFromRule() []*securityv1beta1api.Rule_From {
	return []*securityv1beta1api.Rule_From{
		{
			Source: &securityv1beta1api.Source{
				Namespaces: []string{"istio-gateways"},
			},
		},
	}
}

func getDefaultDenyPolicy(application *skiperatorv1alpha1.Application, denyPaths []string) securityv1beta1.AuthorizationPolicy {
	return securityv1beta1.AuthorizationPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: application.Namespace,
			Name:      application.Name + "-deny",
		},
		Spec: securityv1beta1api.AuthorizationPolicy{
			Action: securityv1beta1api.AuthorizationPolicy_DENY,
			Rules: []*securityv1beta1api.Rule{
				{
					To: []*securityv1beta1api.Rule_To{
						{
							Operation: &securityv1beta1api.Operation{
								Paths: denyPaths,
							},
						},
					},
					From: getGeneralFromRule(),
				},
			},
			Selector: &typev1beta1.WorkloadSelector{
				MatchLabels: util.GetPodAppSelector(application.Name),
			},
		},
	}
}
