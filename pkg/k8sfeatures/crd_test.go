package k8sfeatures

import (
	"context"
	"testing"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsfake "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestIsCRDPresent(t *testing.T) {
	client := apiextensionsfake.NewSimpleClientset(crd("widgets.example.com", "v1"))

	if !IsCRDPresent(context.Background(), client, "widgets.example.com", "v1") {
		t.Fatal("expected CRD with served version to be present")
	}
}

func TestCheckCRDPresentMissingCRD(t *testing.T) {
	client := apiextensionsfake.NewSimpleClientset()

	if err := CheckCRDPresent(context.Background(), client, "widgets.example.com"); err == nil {
		t.Fatal("expected missing CRD to return an error")
	}
}

func TestCheckCRDPresentMissingVersion(t *testing.T) {
	client := apiextensionsfake.NewSimpleClientset(crd("widgets.example.com", "v1alpha1"))

	if err := CheckCRDPresent(context.Background(), client, "widgets.example.com", "v1"); err == nil {
		t.Fatal("expected missing served version to return an error")
	}
}

func crd(name string, versions ...string) *apiextensionsv1.CustomResourceDefinition {
	crdVersions := make([]apiextensionsv1.CustomResourceDefinitionVersion, 0, len(versions))
	for _, version := range versions {
		crdVersions = append(crdVersions, apiextensionsv1.CustomResourceDefinitionVersion{
			Name:    version,
			Served:  true,
			Storage: version == versions[0],
		})
	}

	return &apiextensionsv1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: apiextensionsv1.CustomResourceDefinitionSpec{
			Versions: crdVersions,
		},
	}
}
