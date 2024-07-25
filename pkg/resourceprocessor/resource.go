package resourceprocessor

import (
	"github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/reconciliation"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/resourceutils"
	"github.com/kartverket/skiperator/pkg/util"
	"golang.org/x/exp/maps"
	v1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func copyRequiredData(new client.Object, existing client.Object) {
	new.SetResourceVersion(existing.GetResourceVersion())
	new.SetUID(existing.GetUID())
	new.SetCreationTimestamp(existing.GetCreationTimestamp())
	new.SetSelfLink(existing.GetSelfLink())
	new.SetOwnerReferences(existing.GetOwnerReferences())
}

// Patch if you care about status or if kubernetes does changes to the object after creation
func requirePatch(obj client.Object) bool {
	switch obj.(type) {
	case *v1.Deployment:
		return true
	case *batchv1.Job:
		return true
	}
	return false
}

func preparePatch(new client.Object, old client.Object) {
	switch new.(type) {
	case *v1.Deployment:
		deployment := old.(*v1.Deployment)
		definition := new.(*v1.Deployment)
		if definition.Spec.Replicas == nil {
			definition.Spec.Replicas = deployment.Spec.Replicas
		}
		// The command "kubectl rollout restart" puts an annotation on the deployment template in order to track
		// rollouts of different replicasets. This annotation must not trigger a new reconcile, and a quick and easy
		// fix is to just remove it from the map before hashing and checking the diff.
		if _, rolloutIssued := deployment.Spec.Template.Annotations["kubectl.kubernetes.io/restartedAt"]; rolloutIssued {
			delete(deployment.Spec.Template.Annotations, "kubectl.kubernetes.io/restartedAt")
		}
	case *batchv1.Job:
		job := old.(*batchv1.Job)
		definition := new.(*batchv1.Job)
		maps.Copy(definition.Spec.Template.Labels, job.Spec.Template.Labels) //kubernetes adds labels on creation
		definition.Spec.Selector = job.Spec.Selector                         //is set on creation
	}
}

func diffBetween(old client.Object, new client.Object) bool {
	switch new.(type) {
	case *v1.Deployment:
		deployment := old.(*v1.Deployment)
		definition := new.(*v1.Deployment)
		deploymentHash := util.GetHashForStructs([]interface{}{&deployment.Spec, &deployment.Labels})
		deploymentDefinitionHash := util.GetHashForStructs([]interface{}{&definition.Spec, &definition.Labels})
		if deploymentHash != deploymentDefinitionHash {
			return true
		}

		// Same mechanism as "pod-template-hash"
		if equality.Semantic.DeepEqual(deployment.DeepCopy().Spec, definition.DeepCopy().Spec) {
			return false
		}

		return true

	case *batchv1.Job:
		job := old.(*batchv1.Job)
		definition := new.(*batchv1.Job)
		jobHash := util.GetHashForStructs([]interface{}{&job.Spec, &job.Labels})
		jobDefinitionHash := util.GetHashForStructs([]interface{}{&definition.Spec, &definition.Labels})
		if jobHash != jobDefinitionHash {
			return true
		}
	}
	return true
}

// TODO maybe this should be a reconciliation function instead
func getLabelsForResources(task reconciliation.Reconciliation) map[string]string {
	if task.GetType() == reconciliation.ApplicationType {
		app := task.GetReconciliationObject().(*v1alpha1.Application)
		return resourceutils.GetApplicationDefaultLabels(app)
	} else if task.GetType() == reconciliation.NamespaceType {
		return resourceutils.GetNamespaceLabels()
	} else if task.GetType() == reconciliation.RoutingType {
		routing := task.GetReconciliationObject().(*v1alpha1.Routing)
		return resourceutils.GetRoutingLabels(routing)
	} else if task.GetType() == reconciliation.JobType {
		skipjob := task.GetReconciliationObject().(*v1alpha1.SKIPJob)
		return resourceutils.GetSKIPJobLabels(skipjob)
	}
	return nil
}
