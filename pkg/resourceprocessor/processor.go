package resourceprocessor

import (
	"github.com/kartverket/skiperator/api/v1alpha1"
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

func (r *ResourceProcessor) Process(task reconciliation.Reconciliation) []error {
	if !hasGVK(task.GetResources()) {
		return []error{v1alpha1.ErrNoGVK}
	}
	shouldDelete, shouldUpdate, shouldPatch, shouldCreate, err := r.getDiff(task)
	if err != nil {
		return []error{err}
	}
	results := map[client.Object]error{}

	for _, obj := range shouldDelete {
		err = r.delete(task.GetCtx(), obj)
		results[obj] = err
	}

	for _, obj := range shouldCreate {
		err = r.create(task.GetCtx(), obj)
		results[obj] = err
	}

	for _, obj := range shouldPatch {
		err = r.patch(task.GetCtx(), obj)
		results[obj] = err
	}

	for _, obj := range shouldUpdate {
		err = r.update(task.GetCtx(), obj)
		results[obj] = err
	}

	var errors []error
	for obj, err := range results {
		if err != nil {
			task.GetSKIPObject().GetStatus().AddSubResourceStatus(obj, err.Error(), v1alpha1.ERROR)
			errors = append(errors, err)
		} else {
			task.GetSKIPObject().GetStatus().AddSubResourceStatus(obj, "Resource successfully synced", v1alpha1.SYNCED)
		}
	}
	return errors
}
