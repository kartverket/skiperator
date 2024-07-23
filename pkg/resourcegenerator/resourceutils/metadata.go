package resourceutils

import (
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/job"
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

// TODO Generalize this so we dont need a type
func GetApplicationDefaultLabels(application *skiperatorv1alpha1.Application) map[string]string {
	return map[string]string{
		"app.kubernetes.io/managed-by":            "skiperator",
		"skiperator.skiperator.no/controller":     "application",
		"app":                                     application.Name,
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

func SetNamespaceLabels(object client.Object) {
	labels := object.GetLabels()
	if len(labels) == 0 {
		labels = make(map[string]string)
	}
	maps.Copy(labels, GetNamespaceLabels())
	object.SetLabels(labels)
}

func GetNamespaceLabels() map[string]string {
	return map[string]string{
		"app.kubernetes.io/managed-by":        "skiperator",
		"skiperator.skiperator.no/controller": "namespace",
	}
}

func SetRoutingLabels(object client.Object, routing *skiperatorv1alpha1.Routing) {
	labels := object.GetLabels()
	if len(labels) == 0 {
		labels = make(map[string]string)
	}
	maps.Copy(labels, GetRoutingLabels(routing))
	object.SetLabels(labels)
}

func GetRoutingLabels(routing *skiperatorv1alpha1.Routing) map[string]string {
	return map[string]string{
		"app.kubernetes.io/managed-by":              "skiperator",
		"skiperator.kartverket.no/controller":       "routing",
		"skiperator.kartverket.no/routing-name":     routing.Name,
		"skiperator.kartverket.no/source-namespace": routing.Namespace,
	}
}

// TODO Porbably smart to move these SET functions to the controllers or types
func SetSKIPJobLabels(object client.Object, skipJob *skiperatorv1alpha1.SKIPJob) {
	labels := object.GetLabels()
	if len(labels) == 0 {
		labels = make(map[string]string)
	}
	maps.Copy(labels, GetSKIPJobLabels(skipJob))
	object.SetLabels(labels)
}

// TODO these labels are a disaster
func GetSKIPJobLabels(skipJob *skiperatorv1alpha1.SKIPJob) map[string]string {
	return map[string]string{
		"app":                                 skipJob.KindPostFixedName(),
		"app.kubernetes.io/managed-by":        "skiperator",
		"skiperator.kartverket.no/controller": "skipjob",
		// Used by hahaha to know that the Pod should be watched for killing sidecars
		job.IsSKIPJobKey: "true",
		// Added to be able to add the SKIPJob to a reconcile queue when Watched Jobs are queued
		job.SKIPJobReferenceLabelKey: skipJob.Name,
	}
}
