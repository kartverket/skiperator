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

func GetApplicationDefaultLabels(application *skiperatorv1alpha1.Application) map[string]string {
	return map[string]string{
		"app.kubernetes.io/managed-by":            "skiperator",
		"skiperator.skiperator.no/controller":     "application",
		"application.skiperator.no/app":           application.Name,
		"application.skiperator.no/app-name":      application.Name,
		"application.skiperator.no/app-namespace": application.Namespace,
	}
}

func SetApplicationLabels(object client.Object, app *skiperatorv1alpha1.Application) {
	labels := object.GetLabels()
	if len(labels) == 0 {
		labels = make(map[string]string)
	}
	if app.Spec.Labels != nil {
		maps.Copy(labels, app.Spec.Labels)
	}
	maps.Copy(labels, GetApplicationDefaultLabels(app))
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
		if strings.ToLower(k) == strings.ToLower(resourceKind) {
			return v, true
		}
	}
	return nil, false
}
