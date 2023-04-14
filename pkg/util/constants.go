package util

var CommonAnnotations = map[string]string{
	// Prevents Argo CD from deleting these resources and leaving the namespace
	// in a deadlocked deleting state
	// https://argo-cd.readthedocs.io/en/stable/user-guide/sync-options/#no-prune-resources
	"argocd.argoproj.io/sync-options": "Prune=false",
}
