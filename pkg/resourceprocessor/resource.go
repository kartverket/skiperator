package resourceprocessor

import "sigs.k8s.io/controller-runtime/pkg/client"

func copyRequiredData(new client.Object, existing client.Object) {
	new.SetResourceVersion(existing.GetResourceVersion())
	new.SetUID(existing.GetUID())
	new.SetCreationTimestamp(existing.GetCreationTimestamp())
	new.SetSelfLink(existing.GetSelfLink())
	new.SetOwnerReferences(existing.GetOwnerReferences())
}
