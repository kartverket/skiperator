package controllers

import (
	"golang.org/x/exp/slices"
	corev1 "k8s.io/api/core/v1"
	"regexp"
)

var internalPattern = regexp.MustCompile(`[^.]\.skip\.statkart\.no`)

func isInternal(hostname string) bool {
	return internalPattern.MatchString(hostname)
}

var excludedNamespaces = []string{
	// System namespaces
	"istio-system",
	"kube-node-lease",
	"kube-public",
	"kube-system",
	"skiperator-system",
	"config-management-system",
	"config-management-monitoring",
	"asm-system",
	"anthos-identity-service",
	"binauthz-system",
	"cert-manager",
	"gatekeeper-system",
	"gke-connect",
	"gke-system",
	"resource-group-system",
	// Bundles NetworkPolicies already
	"kasten-io",
	// TODO needs NetworkPolicies/Skiperator
	"vault",
	// TODO needs NetworkPolicies/Skiperator
	"nibas",
	// TODO needs NetworkPolicies/Skiperator
	"aut",
	// TODO PoC, add NetworkPolicies after
	"sysdig-agent",
	"sysdig-admission-controller",
}

func isNotExcludedNamespace(namespace *corev1.Namespace) bool {
	return !slices.Contains(excludedNamespaces, namespace.Name)
}
