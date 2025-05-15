package resourceprocessor

import (
	"fmt"
	"github.com/kartverket/skiperator/pkg/reconciliation"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type resourceDiff struct {
	shouldDelete []client.Object
	shouldUpdate map[runtime.Unstructured]client.Object
	shouldPatch  map[runtime.Unstructured]client.Object
	shouldCreate []client.Object
}

func (r *ResourceProcessor) getDiff(task reconciliation.Reconciliation) (*resourceDiff, error) {
	liveObjects := make([]runtime.Unstructured, 0)
	labels := task.GetSKIPObject().GetDefaultLabels()

	if labels == nil {
		return nil, fmt.Errorf("labels are nil, cant process resources without labels")
	}
	if err := r.listResourcesByLabels(task.GetCtx(), getNamespace(task), labels, &liveObjects); err != nil {
		return nil, fmt.Errorf("failed to list resources by labels: %w", err)
	}
	//TODO ugly as hell
	certs := make([]runtime.Unstructured, 0)
	if err := r.getCertificates(task.GetCtx(), labels, &certs); err != nil {
		return nil, fmt.Errorf("failed to get certificates: %w", err)
	}
	liveObjects = append(liveObjects, certs...)
	liveObjectsMap := make(map[string]runtime.Unstructured)
	for _, obj := range liveObjects {
		liveObjectsMap[client.ObjectKeyFromObject(obj.(client.Object)).String()+obj.GetObjectKind().GroupVersionKind().Kind] = obj
	}

	newObjectsMap := make(map[string]client.Object)
	for _, obj := range task.GetResources() {
		newObjectsMap[client.ObjectKeyFromObject(obj).String()+(obj).GetObjectKind().GroupVersionKind().Kind] = obj
	}

	diffs := &resourceDiff{
		shouldDelete: make([]client.Object, 0),
		shouldUpdate: make(map[runtime.Unstructured]client.Object),
		shouldPatch:  make(map[runtime.Unstructured]client.Object),
		shouldCreate: make([]client.Object, 0),
	}

	// Determine resources to delete
	for key, liveObj := range liveObjectsMap {
		if shouldIgnoreObject(liveObj.(client.Object)) {
			continue
		}
		if _, exists := newObjectsMap[key]; !exists {
			diffs.shouldDelete = append(diffs.shouldDelete, liveObj.(client.Object))
		}
	}

	for key, newObj := range newObjectsMap {
		if liveObj, exists := liveObjectsMap[key]; exists {
			if shouldIgnoreObject(liveObj.(client.Object)) {
				continue
			}
			if compareObject(liveObj.(client.Object), newObj) {
				if requirePatch(newObj) {
					diffs.shouldPatch[liveObj] = newObj
				} else {
					diffs.shouldUpdate[liveObj] = newObj
				}
			}
		} else {
			diffs.shouldCreate = append(diffs.shouldCreate, newObj)
		}
	}

	return diffs, nil
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
		return r.GetSKIPObject().GetName()
	}
	return r.GetSKIPObject().GetNamespace()
}

func shouldIgnoreObject(obj client.Object) bool {
	if obj.GetLabels()["skiperator.kartverket.no/ignore"] == "true" {
		return true
	}
	if len(obj.GetOwnerReferences()) > 0 && obj.GetOwnerReferences()[0].Kind == "CronJob" {
		return true
	}
	return false
}
