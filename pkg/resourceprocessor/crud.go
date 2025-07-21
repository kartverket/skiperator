package resourceprocessor

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"reflect"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *ResourceProcessor) create(ctx context.Context, obj client.Object) error {
	createObj := obj.DeepCopyObject().(client.Object) //copy so we keep gvk
	err := r.client.Create(ctx, createObj)
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
			r.log.Info("Couldn't find object trying to update. Attempting create.", "kind", resource.GetObjectKind().GroupVersionKind().Kind, "name", resource.GetName())
			return r.create(ctx, resource)
		}
		r.log.Error(err, "Failed to get object, for unknown reason")
	}
	identical := isObjectIdentical(resource, existing)
	if identical {
		return nil
	}
	copyRequiredData(resource, existing)
	if err := r.client.Update(ctx, resource); err != nil {
		r.log.Error(err, "Failed to update object")
		return err
	}
	return nil
}

func (r *ResourceProcessor) patch(ctx context.Context, newObj client.Object) error {
	existing := newObj.DeepCopyObject().(client.Object)
	if err := r.client.Get(ctx, client.ObjectKeyFromObject(newObj), existing); err != nil {
		if errors.IsNotFound(err) {
			r.log.Info("Couldn't find object trying to update. Attempting create.", "kind", newObj.GetObjectKind().GroupVersionKind().Kind, "name", newObj.GetName())
			return r.create(ctx, newObj)
		}
		r.log.Error(err, "Failed to get object, for unknown reason")
	}

	preparePatch(newObj, existing)

	identical := isObjectIdentical(newObj, existing)
	if identical {
		r.log.Info("No diff between objects, not patching", "kind", newObj.GetObjectKind().GroupVersionKind().Kind, "name", newObj.GetName())
		return nil
	}

	err := r.client.Patch(ctx, newObj, client.MergeFrom(existing))
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

func (r *ResourceProcessor) getCertificates(ctx context.Context, labels map[string]string, objList *[]client.Object) error {
	return r.listResourcesByLabels(ctx, "istio-gateways", labels, objList)
}

func isObjectIdentical(resource, existing client.Object) bool {
	// Compare Labels
	if !reflect.DeepEqual(resource.GetLabels(), existing.GetLabels()) {
		return false
	}

	// Compare Spec hash
	resourceSpecHash, err := getSpecHash(resource)
	if err != nil || resourceSpecHash == "" {
		return false
	}
	existingSpecHash, err := getSpecHash(existing)
	if err != nil {
		return false
	}
	return resourceSpecHash == existingSpecHash
}

func getSpecHash(obj client.Object) (string, error) {
	val := reflect.ValueOf(obj).Elem().FieldByName("Spec")
	if !val.IsValid() {
		return "", nil
	}
	specBytes, err := json.Marshal(val.Interface())
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256(specBytes)
	return fmt.Sprintf("%x", hash), nil
}
