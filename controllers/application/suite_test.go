/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package applicationcontroller_test

import (
	"context"
	"fmt"
	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	applicationcontroller "github.com/kartverket/skiperator/controllers/application"
	"github.com/kartverket/skiperator/pkg/util"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"os"
	"path/filepath"
	"runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/envtest/komega"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	skiperatorkartverketnov1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	//+kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var (
	cfg       *rest.Config
	k8sClient client.Client
	testEnv   *envtest.Environment
	ctx       context.Context
	cancel    context.CancelFunc
	n         int
)

const (
	testNamespace      = "test"
	otherTestNamespace = "other"
	clusterVersion     = "1.27.1"
)

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Controller Suite")
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))
	Expect(os.Setenv("KUBEBUILDER_ASSETS", fmt.Sprintf("../../bin/k8s/%v-%v-%v", clusterVersion, runtime.GOOS, runtime.GOARCH))).To(Succeed())
	ctx, cancel = context.WithCancel(context.TODO())
	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{
			filepath.Join("..", "..", "config", "crd"),
			// Using
			// Istio v1.17.2
			// Prometheus v0.66.0
			// Cert-Manager v1.12.4
			filepath.Join("..", "..", "config", "crd", "external"),
		},
		ErrorIfCRDPathMissing: true,
	}

	var err error
	// cfg is defined in this file globally.
	cfg, err = testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	Expect(scheme.AddToScheme(scheme.Scheme)).To(Succeed())
	Expect(autoscalingv2.AddToScheme(scheme.Scheme)).To(Succeed())
	Expect(securityv1beta1.AddToScheme(scheme.Scheme)).To(Succeed())
	Expect(networkingv1beta1.AddToScheme(scheme.Scheme)).To(Succeed())
	Expect(certmanagerv1.AddToScheme(scheme.Scheme)).To(Succeed())
	Expect(policyv1.AddToScheme(scheme.Scheme)).To(Succeed())
	Expect(monitoringv1.AddToScheme(scheme.Scheme)).To(Succeed())
	Expect(skiperatorkartverketnov1alpha1.AddToScheme(scheme.Scheme)).To(Succeed())

	//+kubebuilder:scaffold:scheme

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())
	komega.SetClient(k8sClient)
	Expect(err).ToNot(HaveOccurred())

	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		Client: client.Options{Cache: &client.CacheOptions{Unstructured: true}},
		Scheme: scheme.Scheme,
	})
	Expect(err).NotTo(HaveOccurred())

	err = (&applicationcontroller.ApplicationReconciler{
		ReconcilerBase: util.NewFromManager(mgr, mgr.GetEventRecorderFor("application-controller")),
	}).SetupWithManager(mgr)
	Expect(err).ToNot(HaveOccurred())

	go func() {
		defer GinkgoRecover()
		err = mgr.Start(ctx)
		Expect(err).ToNot(HaveOccurred(), "failed to run manager")
	}()

})

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	cancel()
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})

// testEnv doesn't clean up namespaces, so we need a new one for every test
func newNamespace() *corev1.Namespace {
	namespace := &corev1.Namespace{}
	namespace.Name = fmt.Sprintf(`%s-%d`, "test", n)
	Expect(k8sClient.Create(ctx, namespace)).To(Succeed())

	fakePullSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "github-auth",
			Namespace: namespace.Name,
		},
	}
	Expect(k8sClient.Create(ctx, fakePullSecret)).To(Succeed())
	n++
	return namespace
}
