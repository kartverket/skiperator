package applicationcontroller_test

import (
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"istio.io/client-go/pkg/apis/security/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Applications", Ordered, func() {
	var application *skiperatorv1alpha1.Application

	const (
		AppName      = "application"
		AppNamespace = testNamespace

		timeout  = time.Second * 10
		duration = time.Second * 10
		interval = time.Millisecond * 250
	)

	BeforeAll(func() {
		application = &skiperatorv1alpha1.Application{
			ObjectMeta: metav1.ObjectMeta{
				Name:      AppName,
				Namespace: AppNamespace,
			},
			Spec: skiperatorv1alpha1.ApplicationSpec{
				Image: "image",
				Port:  8080,
			},
		}

		Expect(k8sClient.Create(ctx, application)).Should(Succeed())
	})

	AfterAll(func() {
		Expect(k8sClient.Delete(ctx, application)).Should(Succeed())
	})

	Context("When an application is minimal", func() {

		It("should have created a deployment", func() {
			deployment := &appsv1.Deployment{
				ObjectMeta: application.ObjectMeta,
			}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, client.ObjectKeyFromObject(deployment), deployment)
				if err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())
		})

		It("should have created a service account", func() {
			sa := &corev1.ServiceAccount{
				ObjectMeta: application.ObjectMeta,
			}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, client.ObjectKeyFromObject(sa), sa)
				if err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())
		})

		It("should have created an HPA", func() {
			hpa := &autoscalingv2.HorizontalPodAutoscaler{
				ObjectMeta: application.ObjectMeta,
			}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, client.ObjectKeyFromObject(hpa), hpa)
				if err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())
		})

		It("should have created a peerauthentication", func() {
			pa := &v1beta1.PeerAuthentication{
				ObjectMeta: application.ObjectMeta,
			}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, client.ObjectKeyFromObject(pa), pa)
				if err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())
		})

		It("should have created an authorizationpolicy", func() {
			ap := &v1beta1.AuthorizationPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: application.Namespace,
					Name:      application.Name + "-deny",
				},
			}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, client.ObjectKeyFromObject(ap), ap)
				if err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())
		})

		It("should not have created a configmap named foobar", func() {
			cm := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: application.Namespace,
					Name:      "foobar",
				},
			}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, client.ObjectKeyFromObject(cm), cm)

				if errors.IsNotFound(err) {
					return true
				}

				if err != nil {
					return false
				}

				return false
			}, timeout, interval).Should(BeTrue())
		})
	})

})
