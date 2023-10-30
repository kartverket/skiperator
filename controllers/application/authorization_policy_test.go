package applicationcontroller_test

import (
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	securityapi "istio.io/api/security/v1beta1"
	"istio.io/client-go/pkg/apis/security/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

var _ = Describe("AuthorizationPolicy", func() {
	var application *skiperatorv1alpha1.Application

	const (
		AppName      = "application"
		AppNamespace = testNamespace

		timeout  = time.Second * 10
		duration = time.Second * 10
		interval = time.Millisecond * 250
	)

	//Set common config
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

	Context("When an application is minimal", Ordered, func() {
		ap := &v1beta1.AuthorizationPolicy{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: application.Namespace,
				Name:      application.Name + "-deny",
			},
		}

		BeforeAll(func() {
			//Set test specific application values
			Expect(k8sClient.Create(ctx, application)).Should(Succeed())
			Eventually(func() bool {
				err := k8sClient.Get(ctx, client.ObjectKeyFromObject(ap), ap)
				return err == nil
			}, timeout).Should(BeTrue())
		})

		AfterAll(func() {
			Expect(k8sClient.Delete(ctx, application)).Should(Succeed())
		})

		It("should have action deny", func() {
			Expect(ap.Spec.Action).Should(Equal(securityapi.AuthorizationPolicy_DENY))
		})

		It("Should have rule from namespace istio gateways", func() {
			Expect(ap.Spec.Rules[0].From[0].Source.Namespaces[0]).Should(Equal("istio-gateways"))
		})

		It("Should have rule to actuator path", func() {
			Expect(ap.Spec.Rules[0].To[0].Operation.Paths[0]).Should(Equal("/actuator*"))
		})

		It("Should selector", func() {
			Expect(ap.Spec.Selector.MatchLabels["app"]).Should(Equal(application.Name))
		})
	})
})
