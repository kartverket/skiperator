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

	defaultDenyAuthPolicy := getDefaultActuatorDenyPolicy(application)

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

		allowListAuthPolicy := getAllowListPolicy(application)
		if len(application.Spec.AuthorizationSettings.AllowList) > 0 {
			newAllowRule := securityv1beta1api.Rule{
				To:   []*securityv1beta1api.Rule_To{},
				From: getGeneralFromRule(),
			}
			for _, endpoint := range application.Spec.AuthorizationSettings.AllowList {
				newToRule := securityv1beta1api.Rule_To{
					Operation: &securityv1beta1api.Operation{
						Paths: []string{endpoint},
					},
				}

				newAllowRule.To = append(newAllowRule.To, &newToRule)
			}

			_, err := ctrlutil.CreateOrPatch(ctx, r.GetClient(), &allowListAuthPolicy, func() error {
				err := ctrlutil.SetControllerReference(application, &allowListAuthPolicy, r.GetScheme())
				if err != nil {
					r.SetControllerError(ctx, application, controllerName, err)
					return err
				}
				// Reinitialise instead of append to avoid appending rules on AllowList change
				allowListAuthPolicy.Spec.Rules = []*securityv1beta1api.Rule{&newAllowRule}

				r.SetLabelsFromApplication(ctx, &allowListAuthPolicy, *application)
				util.SetCommonAnnotations(&allowListAuthPolicy)

				return nil
			})

			if err != nil {
				r.SetControllerError(ctx, application, controllerName, err)
				return reconcile.Result{}, err
			}
		} else {
			err := r.GetClient().Delete(ctx, &allowListAuthPolicy)
			err = client.IgnoreNotFound(err)
			if err != nil {
				r.SetControllerError(ctx, application, controllerName, err)
				return reconcile.Result{}, err
			}
		}
	}

	_, err := ctrlutil.CreateOrPatch(ctx, r.GetClient(), &defaultDenyAuthPolicy, func() error {
		err := ctrlutil.SetControllerReference(application, &defaultDenyAuthPolicy, r.GetScheme())
		if err != nil {
			r.SetControllerError(ctx, application, controllerName, err)
			return err
		}

		r.SetLabelsFromApplication(ctx, &defaultDenyAuthPolicy, *application)
		util.SetCommonAnnotations(&defaultDenyAuthPolicy)

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

func getDefaultActuatorDenyPolicy(application *skiperatorv1alpha1.Application) securityv1beta1.AuthorizationPolicy {
	return securityv1beta1.AuthorizationPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: application.Namespace,
			Name:      application.Name + "-actuator-deny",
		},
		Spec: securityv1beta1api.AuthorizationPolicy{
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
					From: getGeneralFromRule(),
				},
			},
			Selector: &typev1beta1.WorkloadSelector{
				MatchLabels: util.GetApplicationSelector(application.Name),
			},
		},
	}
}

func getAllowListPolicy(application *skiperatorv1alpha1.Application) securityv1beta1.AuthorizationPolicy {
	return securityv1beta1.AuthorizationPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: application.Namespace,
			Name:      application.Name + "-allow",
		},
		Spec: securityv1beta1api.AuthorizationPolicy{
			Selector: &typev1beta1.WorkloadSelector{
				MatchLabels: util.GetApplicationSelector(application.Name),
			},
			Rules:  []*securityv1beta1api.Rule{},
			Action: securityv1beta1api.AuthorizationPolicy_ALLOW,
		},
	}
}
