package resourceprocessor

import (
	"context"
	"fmt"
	"github.com/kartverket/skiperator/pkg/reconciliation"
	"k8s.io/apimachinery/pkg/api/meta"
	"reflect"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *ResourceProcessor) getDiff(task reconciliation.Reconciliation) ([]*client.Object, error) {
	var liveObjects client.ObjectList
	if err := r.listResourcesByLabels(task.GetCtx(), task.GetReconciliationObject().GetNamespace(), task.GetReconciliationObject().GetLabels(), liveObjects); err != nil {
		return nil, fmt.Errorf("failed to list resources by labels: %w", err)
	}

	diff := make([]*client.Object, 0)
	for _, liveObj := range liveObjects {
		for _, syncObj := range task.GetSyncObjects() {
			if r.compareObject(liveObj, syncObj) {
				// Remove the object from the list of live objects
				// so that we can identify the objects that need to be deleted
				diff = append(diff, liveObj)
			}
		}
	}
	return diff, nil
}

func (r *ResourceProcessor) compareObject(obj1, obj2 client.Object) bool {
	// List doesnt return with group version kind. https://github.com/kubernetes/client-go/issues/308
	obj1Meta, err := meta.Accessor(obj1)
	if err != nil {
		r.log.Error(err, "failed to get object meta", obj1.GetName())
		return true
	}

	obj2Meta, err := meta.Accessor(obj2)
	if err != nil {
		r.log.Error(err, "failed to get object meta", obj2.GetName())
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

func (r *ResourceProcessor) listResourcesByLabels(ctx context.Context, namespace string, labels map[string]string, objList []*client.Object) error {
	// Convert the map of labels to a selector string
	selector := metav1.LabelSelector{MatchLabels: labels}
	selectorString, err := metav1.LabelSelectorAsSelector(&selector)
	if err != nil {
		return fmt.Errorf("failed to convert label selector to selector string: %w", err)
	}

	// List options that include the label selector and namespace
	listOpts := &client.ListOptions{
		LabelSelector: selectorString,
		Namespace:     namespace,
	}

	// List resources
	for _, schema := range r.schemas {
		if err := r.client.List(ctx, schema, listOpts); err != nil {
			return fmt.Errorf("failed to list resources: %w", err)
		}
		for _, resource := range schema.Items {
			objList = append(objList, resource.)
		}
	}


	return nil
}
