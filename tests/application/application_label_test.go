package application_test

import (
	"context"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("ApplicationLabel", func() {
	var (
		application skiperatorv1alpha1.Application
		ns          *corev1.Namespace
	)
	appName := "app"
	label := "test"
	BeforeEach(func() {
		// Initialize applications and namespaces
		ns = newNamespace()

		application = skiperatorv1alpha1.Application{
			ObjectMeta: metav1.ObjectMeta{
				Name:      appName,
				Namespace: ns.Name,
			},
			Spec: skiperatorv1alpha1.ApplicationSpec{
				Image:  "image",
				Port:   8080,
				Labels: map[string]string{"cascadeLabel": label},
			},
		}
	})

	Context("when an application is created", func() {
		BeforeEach(func() {
			ctx := context.Background()
			Expect(k8sClient.Create(ctx, &application)).To(Succeed())
		})

		It("should label the ServiceAccount correctly", func() {
			serviceAccount := &corev1.ServiceAccount{
				ObjectMeta: metav1.ObjectMeta{
					Name:      appName,
					Namespace: ns.Name,
				},
			}
			Eventually(func() bool {
				err := k8sClient.Get(context.Background(), client.ObjectKeyFromObject(serviceAccount), serviceAccount)
				return err == nil
			}, timeout, interval).Should(BeTrue())
			Expect(serviceAccount.Labels).To(HaveKeyWithValue("cascadeLabel", "test"))
		})

		It("should label the Deployment correctly", func() {
			deployment := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      appName,
					Namespace: ns.Name,
				},
			}
			Eventually(func() bool {
				err := k8sClient.Get(context.Background(), client.ObjectKeyFromObject(deployment), deployment)
				return err == nil
			}, timeout, interval).Should(BeTrue())
			Expect(deployment.Labels).To(HaveKeyWithValue("cascadeLabel", "test"))
		})

		It("should label the Application correctly", func() {
			actualApplication := &skiperatorv1alpha1.Application{
				ObjectMeta: metav1.ObjectMeta{
					Name:      appName,
					Namespace: ns.Name,
				},
			}
			Eventually(func() bool {
				err := k8sClient.Get(context.Background(), client.ObjectKeyFromObject(actualApplication), actualApplication)
				return err == nil
			}, timeout, interval).Should(BeTrue())
			Expect(actualApplication.Labels).To(HaveKeyWithValue("cascadeLabel", "test"))
		})
	})

	Context("when an application is deleted", func() {
		BeforeEach(func() {
			ctx := context.Background()
			Expect(k8sClient.Create(ctx, &application)).To(Succeed())
			Expect(k8sClient.Delete(ctx, &application)).To(Succeed())
		})
		It("should remove the application from cluster", func() {
			deletedApplication := &skiperatorv1alpha1.Application{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, client.ObjectKeyFromObject(&application), deletedApplication)
				return err != nil && errors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue())
		})
	})
})
