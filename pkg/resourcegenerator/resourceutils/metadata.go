package resourceutils

import (
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"golang.org/x/exp/maps"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

var commonAnnotations = map[string]string{
	// Prevents Argo CD from deleting these resources and leaving the namespace
	// in a deadlocked deleting state
	// https://argo-cd.readthedocs.io/en/stable/user-guide/sync-options/#no-prune-resources
	"argocd.argoproj.io/sync-options": "Prune=false",
}

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
		object.SetLabels(labels)
	}
	setResourceLabels(object, app)
}

func setResourceLabels(obj client.Object, app *skiperatorv1alpha1.Application) {
	objectGroupVersionKind := obj.GetObjectKind().GroupVersionKind().Kind
	resourceLabels, isPresent := getResourceLabels(app, objectGroupVersionKind)
	if !isPresent {
		return
	}
	maps.Copy(obj.GetLabels(), resourceLabels)
}

func getResourceLabels(app *skiperatorv1alpha1.Application, resourceKind string) (map[string]string, bool) {
	for k, v := range app.Spec.ResourceLabels {
		if strings.ToLower(k) == strings.ToLower(resourceKind) {
			return v, true
		}
	}
	return nil, false
}
