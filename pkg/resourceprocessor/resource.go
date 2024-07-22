package resourceprocessor

import (
	"github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/reconciliation"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/resourceutils"
	"github.com/kartverket/skiperator/pkg/util"
	v1 "k8s.io/api/apps/v1"
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

func requirePatch(obj client.Object) bool {
	switch obj.(type) {
	case *v1.Deployment:
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
	}
	return true
}

func getLabelsForResources(task reconciliation.Reconciliation) map[string]string {
	if task.GetType() == reconciliation.ApplicationType {
		app := task.GetReconciliationObject().(*v1alpha1.Application)
		return resourceutils.GetApplicationDefaultLabels(app)
	} else if task.GetType() == reconciliation.NamespaceType {
		return resourceutils.GetNamespaceLabels()
	}
	return nil
}
