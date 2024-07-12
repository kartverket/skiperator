package resourceprocessor

import (
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *ResourceProcessor) create(ctx context.Context, obj client.Object) error {
	err := r.client.Create(ctx, obj)
	if err != nil && errors.IsAlreadyExists(err) {
		if err = r.update(ctx, obj); err != nil {
			return err
		}
	}
	if err != nil {
		return err
	}
	return nil
}

func (r *ResourceProcessor) update(ctx context.Context, resource client.Object) error {
	existing := resource.DeepCopyObject().(client.Object)
	if err := r.client.Get(ctx, client.ObjectKeyFromObject(resource), existing); err != nil {
		if errors.IsNotFound(err) {
			r.log.Warning("Couldn't find object trying to update. Attempting create.", "kind", resource.GetObjectKind().GroupVersionKind().Kind, "name", resource.GetName())
			return r.create(ctx, resource)
		}
		r.log.Error(err, "Failed to get object, for unknown reason")
	}
	copyRequiredData(resource, existing)
	if err := r.client.Update(ctx, resource); err != nil {
		r.log.Error(err, "Failed to update object")
		return err
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

func (r *ResourceProcessor) listResourcesByLabels(ctx context.Context, namespace string, labels map[string]string, objList *[]client.Object) error {
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
			obj := resource.DeepCopyObject().(client.Object)
			*objList = append(*objList, obj)
		}
	}

	return nil
}
