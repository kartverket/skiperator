package application_test

import (
	"context"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/api/v1alpha1/podtypes"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("ApplicationPorts", func() {
	var (
		application skiperatorv1alpha1.Application
		ns          *corev1.Namespace
	)
	appName := "app"

	BeforeEach(func() {
		// Initialize applications and namespaces
		ns = newNamespace()

		application = skiperatorv1alpha1.Application{
			ObjectMeta: metav1.ObjectMeta{
				Name:      appName,
				Namespace: ns.Name,
			},
			Spec: skiperatorv1alpha1.ApplicationSpec{
				Image: "image",
				Port:  8080,
				AdditionalPorts: []podtypes.InternalPort{
					{
						Name:     "metrics",
						Port:     8181,
						Protocol: "TCP",
					},
					{
						Name:     "some-udp-port",
						Port:     8282,
						Protocol: "UDP",
					},
				},
			},
		}
	})

	Context("when an application is created", func() {
		BeforeEach(func() {
			ctx := context.Background()
			Expect(k8sClient.Create(ctx, &application)).To(Succeed())
		})

		It("should create deployment with additional ports", func() {
			deployment := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      appName,
					Namespace: ns.Name,
				},
			}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, client.ObjectKeyFromObject(deployment), deployment)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			By("checking the deployment ports")
			container := deployment.Spec.Template.Spec.Containers[0]
			Expect(container).NotTo(BeNil())
			Expect(container.Ports).To(ConsistOf(
				corev1.ContainerPort{Name: "main", ContainerPort: 8080, Protocol: corev1.ProtocolTCP},
				corev1.ContainerPort{Name: "metrics", ContainerPort: 8181, Protocol: corev1.ProtocolTCP},
				corev1.ContainerPort{Name: "some-udp-port", ContainerPort: 8282, Protocol: corev1.ProtocolUDP},
			))
		})

		It("should create a service with the additional ports", func() {
			service := &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      appName,
					Namespace: ns.Name,
				},
			}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, client.ObjectKeyFromObject(service), service)
				return err == nil
			}, timeout, interval).Should(BeTrue())
			By("checking the service ports")
			Expect(service.Spec.Ports).To(ConsistOf(
				corev1.ServicePort{Name: "metrics", Port: 8181, TargetPort: intstr.FromInt(8181), Protocol: corev1.ProtocolTCP},
				corev1.ServicePort{Name: "some-udp-port", Port: 8282, TargetPort: intstr.FromInt(8282), Protocol: corev1.ProtocolUDP},
				corev1.ServicePort{Name: "http", Port: 8080, TargetPort: intstr.FromInt(8080), Protocol: corev1.ProtocolTCP, AppProtocol: pointer.StringPtr("http")},
			))
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
