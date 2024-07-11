package controllers

import (
	"context"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func getIdentityConfigMap(client client.Client) (*corev1.ConfigMap, error) {
	namespacedName := types.NamespacedName{Name: "gcp-identity-config", Namespace: "skiperator-system"}
	identityConfigMap := &corev1.ConfigMap{}
	if err := client.Get(context.Background(), namespacedName, identityConfigMap); err != nil {
		return nil, err
	}
	return identityConfigMap, nil
}
