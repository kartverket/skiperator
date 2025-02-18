package resourceprocessor

import (
	"github.com/kartverket/skiperator/pkg/util"
	nais_io_v1 "github.com/nais/liberator/pkg/apis/nais.io/v1"
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
	case *nais_io_v1.Jwker:
		return true
	}
	return false
}

func preparePatch(new client.Object, old client.Object) {
	switch new.(type) {
	case *v1.Deployment:
		deployment := old.(*v1.Deployment)
		definition := new.(*v1.Deployment)

		// Handling HPA.
		// If the replicas field is not set in the definition, we should set it to 1 if the deployment has 0 replicas or HPA will not work.
		// If the replicas field is set in the definition, we should use that value so we don't scale down.
		if definition.Spec.Replicas == nil {
			definition.Spec.Replicas = deployment.Spec.Replicas

			if *deployment.Spec.Replicas == int32(0) {
				definition.Spec.Replicas = util.PointTo(int32(1))
			}
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
		maps.Copy(definition.Spec.Template.Labels, job.Spec.Template.Labels) // kubernetes adds labels on creation
		definition.Spec.Selector = job.Spec.Selector                         // immutable
		definition.Spec.Template = job.Spec.Template                         // immutable
		definition.Spec.Completions = job.Spec.Completions                   // immutable
	case *nais_io_v1.Jwker:
		copyRequiredData(new, old)
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
	case *nais_io_v1.Jwker:
		jwker := old.(*nais_io_v1.Jwker)
		definition := new.(*nais_io_v1.Jwker)
		jwkerHash := util.GetHashForStructs([]interface{}{&jwker.Spec})
		definitionHash := util.GetHashForStructs([]interface{}{&definition.Spec})
		return jwkerHash != definitionHash
	}

	return true
}

func hasGVK(resources []client.Object) bool {
	for _, obj := range resources {
		gvk := (obj).GetObjectKind().GroupVersionKind().Kind
		if gvk == "" {
			return false
		}
	}
	return true
}
