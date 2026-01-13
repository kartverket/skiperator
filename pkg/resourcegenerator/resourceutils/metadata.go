package resourceutils

import (
	"maps"
	"strings"

	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	skiperatorv1beta1 "github.com/kartverket/skiperator/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	commonAnnotations = map[string]string{
		// Prevents Argo CD from deleting these resources and leaving the namespace
		// in a deadlocked deleting state
		// https://argo-cd.readthedocs.io/en/stable/user-guide/sync-options/#no-prune-resources
		"argocd.argoproj.io/sync-options": "Prune=false",
	}
)

func SetCommonAnnotations(object client.Object) {
	annotations := object.GetAnnotations()
	if len(annotations) == 0 {
		annotations = make(map[string]string)
	}
	maps.Copy(annotations, commonAnnotations)
	object.SetAnnotations(annotations)
}

func SetApplicationLabels(object client.Object, app *skiperatorv1alpha1.Application) {
	labels := object.GetLabels()
	if len(labels) == 0 {
		labels = make(map[string]string)
	}
	if app.Spec.Labels != nil {
		maps.Copy(labels, app.Spec.Labels)
	}
	maps.Copy(labels, app.GetDefaultLabels())
	object.SetLabels(labels)

	setResourceLabels(object, app)
}

func setResourceLabels(obj client.Object, app *skiperatorv1alpha1.Application) {
	objectGroupVersionKind := obj.GetObjectKind().GroupVersionKind().Kind
	resourceLabels, isPresent := getResourceLabels(app, objectGroupVersionKind)
	if !isPresent {
		return
	}
	labels := obj.GetLabels()
	maps.Copy(labels, resourceLabels)
	obj.SetLabels(labels)
}

func getResourceLabels(app *skiperatorv1alpha1.Application, resourceKind string) (map[string]string, bool) {
	for k, v := range app.Spec.ResourceLabels {
		if strings.EqualFold(k, resourceKind) {
			return v, true
		}
	}
	return nil, false
}

func FindResourceLabelErrors(app *skiperatorv1alpha1.Application, resources []client.Object) map[string]map[string]string {
	labelsWithNoMatch := app.Spec.ResourceLabels
	for k := range labelsWithNoMatch {
		for _, resource := range resources {
			if strings.EqualFold(k, resource.GetObjectKind().GroupVersionKind().Kind) {
				delete(labelsWithNoMatch, k)
			}
		}
	}
	return labelsWithNoMatch
}

func SetNamespaceLabels(object client.Object, skipns *skiperatorv1alpha1.SKIPNamespace) {
	labels := object.GetLabels()
	if len(labels) == 0 {
		labels = make(map[string]string)
	}
	maps.Copy(labels, skipns.GetDefaultLabels())
	object.SetLabels(labels)
}

func SetRoutingLabels(object client.Object, routing *skiperatorv1alpha1.Routing) {
	labels := object.GetLabels()
	if len(labels) == 0 {
		labels = make(map[string]string)
	}
	maps.Copy(labels, routing.GetDefaultLabels())
	object.SetLabels(labels)
}

// TODO Porbably smart to move these SET functions to the controllers or types
func SetSKIPJobLabels(object client.Object, skipJob *skiperatorv1beta1.SKIPJob) {
	labels := object.GetLabels()
	if len(labels) == 0 {
		labels = make(map[string]string)
	}
	maps.Copy(labels, skipJob.GetDefaultLabels())
	object.SetLabels(labels)
}
