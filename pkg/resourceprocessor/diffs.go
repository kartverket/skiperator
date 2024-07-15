package resourceprocessor

import (
	"fmt"
	"github.com/kartverket/skiperator/pkg/reconciliation"
	"k8s.io/apimachinery/pkg/api/meta"
	"reflect"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

// TODO fix pointer mess ? ? ? ?
func (r *ResourceProcessor) getDiff(task reconciliation.Reconciliation) ([]client.Object, []client.Object, []client.Object, error) {
	liveObjects := make([]client.Object, 0)

	if err := r.listResourcesByLabels(task.GetCtx(), getNamespace(task), task.GetReconciliationObject().GetLabels(), &liveObjects); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to list resources by labels: %w", err)
	}

	liveObjectsMap := make(map[string]*client.Object)
	for _, obj := range liveObjects {
		liveObjectsMap[client.ObjectKeyFromObject(obj).String()+obj.GetObjectKind().GroupVersionKind().Kind] = &obj
	}

	newObjectsMap := make(map[string]*client.Object)
	for _, obj := range task.GetResources() {
		
		newObjectsMap[client.ObjectKeyFromObject(*obj).String()+meta.GetResourceVersion()] = obj
	}

	shouldDelete := make([]client.Object, 0)
	shouldUpdate := make([]client.Object, 0)
	shouldCreate := make([]client.Object, 0)

	for key, newObj := range newObjectsMap {
		if liveObj, exists := liveObjectsMap[key]; exists {
			if r.compareObject(*liveObj, *newObj) {
				shouldUpdate = append(shouldUpdate, *newObj)
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

	return shouldDelete, shouldUpdate, shouldCreate, nil
}

func (r *ResourceProcessor) compareObject(obj1, obj2 client.Object) bool {
	// List doesnt return with group version kind. https://github.com/kubernetes/client-go/issues/308
	obj1Meta, err := meta.Accessor(obj1)
	if err != nil {
		r.log.Error(err, "failed to get object meta", "name", obj1.GetName())
		return true
	}

	obj2Meta, err := meta.Accessor(obj2)
	if err != nil {
		r.log.Error(err, "failed to get object meta", "name", obj2.GetName())
		return true
	}

	if reflect.TypeOf(obj1) != reflect.TypeOf(obj2) {
		return false
	}

	if obj1Meta.GetNamespace() != obj2Meta.GetNamespace() {
		return false
	}

	if obj1Meta.GetName() != obj2Meta.GetName() {
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
