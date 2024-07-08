package pdb

import (
	"context"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/k8sfeatures"
	"github.com/kartverket/skiperator/pkg/log"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/resourceutils"
	"github.com/kartverket/skiperator/pkg/util"
	policyv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func Generate(ctx context.Context, application *skiperatorv1alpha1.Application) (*policyv1.PodDisruptionBudget, error) {
	ctxLog := log.FromContext(ctx)
	ctxLog.Debug("Attempting to generate pdb for application", application.Name)

	pdb := policyv1.PodDisruptionBudget{ObjectMeta: metav1.ObjectMeta{Namespace: application.Namespace, Name: application.Name}}

	if *application.Spec.EnablePDB {

		resourceutils.SetApplicationLabels(&pdb, application)
		resourceutils.SetCommonAnnotations(&pdb)
		var minReplicas uint

		replicas, err := skiperatorv1alpha1.GetStaticReplicas(application.Spec.Replicas)
		if err != nil {
			replicasStruct, err := skiperatorv1alpha1.GetScalingReplicas(application.Spec.Replicas)
			if err != nil {
				ctxLog.Error(err, "Failed to get replicas")
				return nil, err
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
			MinAvailable: determineMinAvailable(minReplicas),
		}

		if k8sfeatures.EnhancedPDBAvailable() {
			pdb.Spec.UnhealthyPodEvictionPolicy = util.PointTo(policyv1.AlwaysAllow)
		}
		return &pdb, nil
	}

	return nil, nil
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
