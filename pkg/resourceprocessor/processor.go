package resourceprocessor

import (
	"context"
	"github.com/kartverket/skiperator/pkg/log"
	"github.com/kartverket/skiperator/pkg/reconciliation"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Processor interface {
	Process() error
}

type ResourceProcessor struct {
	client  client.Client
	log     log.Logger
	schemas []client.ObjectList
}

func NewResourceProcessor(client client.Client, schemas []client.ObjectList) *ResourceProcessor {
	l := log.FromContext(context.Background()).WithName("ResourceProcessor")
	return &ResourceProcessor{client: client, log: l, schemas: schemas}
}

func (r *ResourceProcessor) Process(task reconciliation.Reconciliation) error {
	shouldDelete, shouldCreate, shouldUpdate, err := r.getDiff(task)

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

	for _, obj := range shouldUpdate {
		if err = r.update(task.GetCtx(), obj); err != nil {
			r.log.Error(err, "Failed to update object")
			return err
		}
	}
	return nil
}
