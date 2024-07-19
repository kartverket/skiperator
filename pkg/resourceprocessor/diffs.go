package resourceprocessor

import (
	"fmt"
	"github.com/kartverket/skiperator/pkg/reconciliation"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// TODO fix pointer mess ? ? ? ?
func (r *ResourceProcessor) getDiff(task reconciliation.Reconciliation) ([]client.Object, []client.Object, []client.Object, []client.Object, error) {
	liveObjects := make([]client.Object, 0)
	//TODO labels to get the resources by should be its own get function
	labels := task.GetReconciliationObject().GetLabels()

	if labels == nil {
		return nil, nil, nil, nil, fmt.Errorf("labels are nil, cant process resources without labels")
	}
	if err := r.listResourcesByLabels(task.GetCtx(), getNamespace(task), labels, &liveObjects); err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to list resources by labels: %w", err)
	}

	liveObjectsMap := make(map[string]*client.Object)
	for _, obj := range liveObjects {
		liveObjectsMap[client.ObjectKeyFromObject(obj).String()+obj.GetObjectKind().GroupVersionKind().Kind] = &obj
	}

	newObjectsMap := make(map[string]*client.Object)
	for _, obj := range task.GetResources() {
		newObjectsMap[client.ObjectKeyFromObject(*obj).String()+(*obj).GetObjectKind().GroupVersionKind().Kind] = obj
	}

	shouldDelete := make([]client.Object, 0)
	shouldUpdate := make([]client.Object, 0)
	shouldPatch := make([]client.Object, 0)
	shouldCreate := make([]client.Object, 0)

	for key, newObj := range newObjectsMap {
		if liveObj, exists := liveObjectsMap[key]; exists {
			should := (*liveObj).GetLabels()["skiperator.kartverket.no/ignore"] != "true"
			if !should {
				continue
			}
			if compareObject(*liveObj, *newObj) {
				if requirePatch(*newObj) {
					shouldPatch = append(shouldPatch, *newObj)
				} else {
					shouldUpdate = append(shouldUpdate, *newObj)
				}
			}
		} else {
			shouldCreate = append(shouldCreate, *newObj)
		}
	}

	// Determine resources to delete
	for key, liveObj := range liveObjectsMap {
		if _, exists := newObjectsMap[key]; !exists {
			shouldDelete = append(shouldDelete, *liveObj)
		}
	}

	return shouldDelete, shouldUpdate, shouldPatch, shouldCreate, nil
}

func compareObject(obj1, obj2 client.Object) bool {
	if obj1.GetObjectKind().GroupVersionKind().Kind != obj2.GetObjectKind().GroupVersionKind().Kind {
		return false
	}
	if obj1.GetObjectKind().GroupVersionKind().Group != obj2.GetObjectKind().GroupVersionKind().Group {
		return false
	}
	if obj1.GetObjectKind().GroupVersionKind().Version != obj2.GetObjectKind().GroupVersionKind().Version {
		return false
	}

	if obj1.GetNamespace() != obj2.GetNamespace() {
		return false
	}

	if obj1.GetName() != obj2.GetName() {
		return false
	}

	return true
}

func getNamespace(r reconciliation.Reconciliation) string {
	if r.GetType() == reconciliation.NamespaceType {
		return r.GetReconciliationObject().GetName()
	}
	return r.GetReconciliationObject().GetNamespace()
}
