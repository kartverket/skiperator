package resourceprocessor

import (
	"github.com/kartverket/skiperator/pkg/log"
	"github.com/kartverket/skiperator/pkg/reconciliation"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Processor interface {
	Process() error
}

type ResourceProcessor struct {
	client  client.Client
	log     log.Logger
	schemas []*unstructured.UnstructuredList
}

func NewResourceProcessor(client client.Client, log log.Logger, schemas []client.ObjectList) *ResourceProcessor {
	return &ResourceProcessor{client: client, log: log, schemas: schemas}
}

func (r *ResourceProcessor) Process(task *reconciliation.Reconciliation) error {
	diff, err := getDiff(task)

	if err != nil {
		return err
	}

	if err = r.delete(diff); err != nil {
		return err
	}

	if err = r.apply(task.GetSyncObjects()); err != nil {
		return err
	}
}
