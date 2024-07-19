package resourceprocessor

import (
	"github.com/kartverket/skiperator/pkg/log"
	"github.com/kartverket/skiperator/pkg/reconciliation"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Processor interface {
	Process() error
}

type ResourceProcessor struct {
	client  client.Client
	log     log.Logger
	schemas []unstructured.UnstructuredList
	scheme  *runtime.Scheme
}

func NewResourceProcessor(client client.Client, schemas []unstructured.UnstructuredList, scheme *runtime.Scheme) *ResourceProcessor {
	l := log.NewLogger().WithName("ResourceProcessor")
	return &ResourceProcessor{client: client, log: l, schemas: schemas, scheme: scheme}
}

func (r *ResourceProcessor) Process(task reconciliation.Reconciliation) error {
	shouldDelete, shouldUpdate, shouldPatch, shouldCreate, err := r.getDiff(task)
	if err != nil {
		return err
	}

	for _, obj := range shouldDelete {
		if err = r.delete(task.GetCtx(), obj); err != nil {
			r.log.Error(err, "Failed to delete object")
			return err
		}
	}

	for _, obj := range shouldCreate {
		if err = r.create(task.GetCtx(), obj); err != nil {
			r.log.Error(err, "Failed to create object")
			return err
		}
	}

	for _, newObj := range shouldPatch {
		if err = r.patch(task.GetCtx(), newObj); err != nil {
			r.log.Error(err, "Failed to patch object")
			return err
		}
	}

	for _, obj := range shouldUpdate {
		if err = r.update(task.GetCtx(), obj); err != nil {
			r.log.Error(err, "Failed to update object")
			return err
		}
	}
	return nil
}

func (r *ResourceProcessor) setMeta(new client.Object, old client.Object) {
	new.SetResourceVersion(old.GetResourceVersion())
	new.SetUID(old.GetUID())
	new.SetSelfLink(old.GetSelfLink())

	existingReferences := old.GetOwnerReferences()

	if len(existingReferences) > 0 {
		new.SetOwnerReferences(existingReferences)
	}
}