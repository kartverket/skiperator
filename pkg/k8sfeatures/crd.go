package k8sfeatures

import (
	"context"
	"fmt"
	"strings"

	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type CRDRequirement struct {
	Name     string
	Versions []string
}

func IsCRDPresent(ctx context.Context, client apiextensionsclient.Interface, name string, versions ...string) bool {
	return CheckCRDPresent(ctx, client, name, versions...) == nil
}

func CheckCRDPresent(ctx context.Context, client apiextensionsclient.Interface, name string, versions ...string) error {
	if client == nil {
		return fmt.Errorf("could not create API extensions client")
	}

	crd, err := client.ApiextensionsV1().CustomResourceDefinitions().Get(
		ctx,
		name,
		metav1.GetOptions{},
	)
	if err != nil {
		return fmt.Errorf("missing CRD %q: %w", name, err)
	}
	if crd == nil {
		return fmt.Errorf("missing CRD %q", name)
	}

	missingVersions := make([]string, 0)
	for _, requiredVersion := range versions {
		found := false
		for _, version := range crd.Spec.Versions {
			if version.Name == requiredVersion && version.Served {
				found = true
				break
			}
		}
		if !found {
			missingVersions = append(missingVersions, requiredVersion)
		}
	}
	if len(missingVersions) > 0 {
		return fmt.Errorf(
			"CRD %q is missing served versions: %s",
			name,
			strings.Join(missingVersions, ", "),
		)
	}

	return nil
}

func CheckCRDsPresent(ctx context.Context, client apiextensionsclient.Interface, requirements ...CRDRequirement) error {
	for _, requirement := range requirements {
		if err := CheckCRDPresent(
			ctx,
			client,
			requirement.Name,
			requirement.Versions...,
		); err != nil {
			return err
		}
	}

	return nil
}
