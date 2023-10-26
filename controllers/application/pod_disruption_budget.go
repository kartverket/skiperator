package applicationcontroller

import (
	"context"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/util"
	policyv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *ApplicationReconciler) reconcilePodDisruptionBudget(ctx context.Context, application *skiperatorv1alpha1.Application) (reconcile.Result, error) {
	controllerName := "PodDisruptionBudget"
	_, _ = r.SetControllerProgressing(ctx, application, controllerName)

	pdb := policyv1.PodDisruptionBudget{ObjectMeta: metav1.ObjectMeta{Namespace: application.Namespace, Name: application.Name}}
	shouldReconcile, err := r.ShouldReconcile(ctx, &pdb)
	if err != nil || !shouldReconcile {
		r.SetControllerFinishedOutcome(ctx, application, controllerName, err)
		return reconcile.Result{}, err
	}

	if *application.Spec.EnablePDB {
		_, err := ctrlutil.CreateOrPatch(ctx, r.GetClient(), &pdb, func() error {
			// Set application as owner of the PDB
			err := ctrlutil.SetControllerReference(application, &pdb, r.GetScheme())
			if err != nil {
				_, _ = r.SetControllerError(ctx, application, controllerName, err)
				return err
			}

			r.SetLabelsFromApplication(&pdb, *application)
			util.SetCommonAnnotations(&pdb)
			var minReplicas uint

			replicas, err := skiperatorv1alpha1.GetStaticReplicas(application.Spec.Replicas)
			if err != nil {
				replicasStruct, err := skiperatorv1alpha1.GetScalingReplicas(application.Spec.Replicas)
				if err != nil {
					r.SetControllerError(ctx, application, controllerName, err)
					return err
				} else {
					minReplicas = replicasStruct.Min
				}
			} else {
				minReplicas = replicas
			}

			pdb.Spec = policyv1.PodDisruptionBudgetSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: util.GetPodAppSelector(application.Name),
				},
				MinAvailable:               determineMinAvailable(minReplicas),
				UnhealthyPodEvictionPolicy: util.PointTo(policyv1.AlwaysAllow),
			}

			return nil
		})

		_, _ = r.SetControllerFinishedOutcome(ctx, application, controllerName, err)
		return reconcile.Result{}, err
	} else {
		err := r.GetClient().Delete(ctx, &pdb)
		err = client.IgnoreNotFound(err)
		if err != nil {
			r.SetControllerError(ctx, application, controllerName, err)
			return reconcile.Result{}, err
		}

		_, _ = r.SetControllerFinishedOutcome(ctx, application, controllerName, err)
		return reconcile.Result{}, nil
	}
}

func determineMinAvailable(replicasAvailable uint) *intstr.IntOrString {
	var value intstr.IntOrString

	if replicasAvailable > 1 {
		value = intstr.FromString("50%")
	} else {
		intstr.FromInt(0)
	}

	return &value
}
