package resourceprocessor

import (
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/runtime"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *ResourceProcessor) create(ctx context.Context, obj client.Object) error {
	createObj := obj.DeepCopyObject().(client.Object) //copy so we keep gvk
	err := r.client.Create(ctx, createObj)
	if err != nil {
		return err
	}
	return nil
}

func (r *ResourceProcessor) update(ctx context.Context, resource client.Object, existing runtime.Unstructured) error {
	copyRequiredData(resource, existing)
	if err := r.client.Update(ctx, resource); err != nil {
		r.log.Error(err, "Failed to update object")
		return err
	}
	return nil
}

func (r *ResourceProcessor) patch(ctx context.Context, newObj client.Object, existing runtime.Unstructured) error {
	preparePatch(newObj, existing)

	//TODO move this to getDiffs?
	if !diffBetween(existing, newObj) {
		r.log.Info("No diff between objects, not patching", "kind", newObj.GetObjectKind().GroupVersionKind().Kind, "name", newObj.GetName())
		return nil
	}
	p := client.MergeFrom(existing.(client.Object))
	err := r.client.Patch(ctx, newObj, p)
	if err != nil {
		return fmt.Errorf("failed to patch object: %w", err)
	}
	return nil
}

func (r *ResourceProcessor) delete(ctx context.Context, resource client.Object) error {
	err := r.client.Delete(ctx, resource)
	if err != nil && errors.IsNotFound(err) {
		return nil
	}
	return err
}

func (r *ResourceProcessor) listResourcesByLabels(ctx context.Context, namespace string, labels map[string]string, objList *[]runtime.Unstructured) error {
	selector := metav1.LabelSelector{MatchLabels: labels}
	selectorString, err := metav1.LabelSelectorAsSelector(&selector)
	if err != nil {
		return fmt.Errorf("failed to convert label selector to selector string: %w", err)
	}

	listOpts := &client.ListOptions{
		LabelSelector: selectorString,
		Namespace:     namespace,
	}

	for _, schema := range r.schemas {
		if err := r.client.List(ctx, &schema, listOpts); err != nil {
			return fmt.Errorf("failed to list resources: %w", err)
		}
		for _, resource := range schema.Items {
			*objList = append(*objList, &resource)
		}
	}

	return nil
}

func (r *ResourceProcessor) getCertificates(ctx context.Context, labels map[string]string, objList *[]runtime.Unstructured) error {
	return r.listResourcesByLabels(ctx, "istio-gateways", labels, objList)
}
