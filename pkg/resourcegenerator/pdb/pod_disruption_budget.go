package pdb

import (
	"fmt"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/k8sfeatures"
	"github.com/kartverket/skiperator/pkg/reconciliation"
	"github.com/kartverket/skiperator/pkg/util"
	policyv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func Generate(r reconciliation.Reconciliation) error {
	ctxLog := r.GetLogger()
	if r.GetType() != reconciliation.ApplicationType {
		return fmt.Errorf("unsupported type %s in pod disruption budget", r.GetType())
	}
	application, ok := r.GetSKIPObject().(*skiperatorv1alpha1.Application)
	if !ok {
		err := fmt.Errorf("failed to cast resource to application")
		ctxLog.Error(err, "Failed to generate pod disruption budget")
		return err
	}
	ctxLog.Debug("Attempting to generate pdb for application", "application", application.Name)

	pdb := policyv1.PodDisruptionBudget{ObjectMeta: metav1.ObjectMeta{Namespace: application.Namespace, Name: application.Name}}

	if *application.Spec.EnablePDB {
		var minReplicas uint

		replicas, err := skiperatorv1alpha1.GetStaticReplicas(application.Spec.Replicas)
		if err != nil {
			replicasStruct, err := skiperatorv1alpha1.GetScalingReplicas(application.Spec.Replicas)
			if err != nil {
				ctxLog.Error(err, "Failed to get replicas")
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
			MinAvailable: determineMinAvailable(minReplicas),
		}

		if k8sfeatures.EnhancedPDBAvailable() {
			pdb.Spec.UnhealthyPodEvictionPolicy = util.PointTo(policyv1.AlwaysAllow)
		}
		var obj client.Object = &pdb
		r.AddResource(&obj)
	}

	return nil
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
