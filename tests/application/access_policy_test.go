package application_test

import (
	"context"
	"github.com/google/go-cmp/cmp/cmpopts"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/api/v1alpha1/podtypes"
	. "github.com/kartverket/skiperator/pkg/util"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	networkingv1beta1api "istio.io/api/networking/v1beta1"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	v1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("AccessPolicy", func() {
	var application skiperatorv1alpha1.Application
	Context("when an application is created", func() {
		It("should create network policy and service entry", func() {
			appName := "access-policy"
			appMinimal := "access-policy-two"
			appOtherMinimal := "access-policy-other"
			ns := newNamespace()
			otherNs := newNamespace()
			application = skiperatorv1alpha1.Application{
				ObjectMeta: metav1.ObjectMeta{
					Name:      appName,
					Namespace: ns.Name,
				},
				Spec: skiperatorv1alpha1.ApplicationSpec{
					Image: "image",
					Port:  8080,
				},
			}
			application.Namespace = ns.Name
			applicationMinimal := application
			applicationOtherMinimal := application
			applicationOtherMinimal.Name = appOtherMinimal
			applicationOtherMinimal.Namespace = otherNs.Name
			applicationMinimal.Name = appMinimal
			applicationMinimal.Namespace = ns.Name
			application.Name = appName

			accessPolicy := podtypes.AccessPolicy{
				Inbound: &podtypes.InboundPolicy{
					Rules: []podtypes.InternalRule{{
						Namespace:   otherNs.Name,
						Application: appOtherMinimal,
					}},
				},
				Outbound: podtypes.OutboundPolicy{
					Rules: []podtypes.InternalRule{{
						Namespace:   otherNs.Name,
						Application: appOtherMinimal,
					}, {
						Application: appMinimal,
					}},
					External: []podtypes.ExternalRule{{
						Host: "example.com",
						Ports: []podtypes.ExternalPort{{
							Name:     "http",
							Port:     80,
							Protocol: "HTTP",
						}}},
						{
							Host: "foo.com",
						}},
				},
			}
			application.Spec.AccessPolicy = &accessPolicy
			ctx := context.Background()
			Expect(k8sClient.Create(ctx, &applicationMinimal)).ShouldNot(HaveOccurred())
			Expect(k8sClient.Create(ctx, &applicationOtherMinimal)).ShouldNot(HaveOccurred())
			Expect(k8sClient.Create(ctx, &application)).ShouldNot(HaveOccurred())

			By("Checking network policy")
			np := &v1.NetworkPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: application.Namespace,
					Name:      appName,
				},
			}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, client.ObjectKeyFromObject(np), np)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			Expect(np.Spec.PodSelector.MatchLabels["app"]).Should(Equal(appName))
			Expect(np.Spec.Ingress[0].From[0].NamespaceSelector.MatchLabels["kubernetes.io/metadata.name"]).Should(Equal(otherNs.Name))
			Expect(np.Spec.Ingress[0].From[0].PodSelector.MatchLabels["app"]).Should(Equal(appOtherMinimal))
			Expect(np.Spec.Egress[0].To[0].NamespaceSelector.MatchLabels["kubernetes.io/metadata.name"]).Should(Equal(otherNs.Name))
			Expect(np.Spec.Egress[0].To[0].PodSelector.MatchLabels["app"]).Should(Equal(appOtherMinimal))
			Expect(&np.Spec.Egress[0].Ports[0].Port.IntVal).Should(Equal(PointTo(int32(8080))))
			Expect(np.Spec.Egress[1].To[0].NamespaceSelector.MatchLabels["kubernetes.io/metadata.name"]).Should(Equal(ns.Name))
			Expect(np.Spec.Egress[1].To[0].PodSelector.MatchLabels["app"]).Should(Equal(appMinimal))
			Expect(&np.Spec.Egress[1].Ports[0].Port.IntVal).Should(Equal(PointTo(int32(8080))))

			By("Checking that two service entries are created")
			actualSeWithPorts := &networkingv1beta1.ServiceEntry{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "access-policy-egress-56cd7aa901014e78",
					Namespace: ns.Name,
				},
			}
			expectedSeWithPortsSpec := networkingv1beta1api.ServiceEntry{
				Hosts:      []string{"example.com"},
				Ports:      []*networkingv1beta1api.ServicePort{{Name: "http", Number: uint32(80), Protocol: "HTTP"}},
				Resolution: networkingv1beta1api.ServiceEntry_DNS,
				ExportTo:   []string{".", "istio-system", "istio-gateways"},
			}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, client.ObjectKeyFromObject(actualSeWithPorts), actualSeWithPorts)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			Expect(actualSeWithPorts.Spec).Should(BeComparableTo(expectedSeWithPortsSpec, cmpopts.IgnoreUnexported(networkingv1beta1api.ServiceEntry{}, networkingv1beta1api.ServicePort{})))

			actualSeJustHost := &networkingv1beta1.ServiceEntry{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "access-policy-egress-3a90cb5d70dc06a",
					Namespace: ns.Name,
				},
			}

			expectedSeJustHost := networkingv1beta1api.ServiceEntry{
				Hosts:      []string{"foo.com"},
				Ports:      []*networkingv1beta1api.ServicePort{{Name: "https", Number: uint32(443), Protocol: "HTTPS"}},
				Resolution: networkingv1beta1api.ServiceEntry_DNS,
				ExportTo:   []string{".", "istio-system", "istio-gateways"},
			}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, client.ObjectKeyFromObject(actualSeJustHost), actualSeJustHost)
				return err == nil
			}, timeout, interval).Should(BeTrue())
			Expect(actualSeJustHost.Spec).Should(BeComparableTo(expectedSeJustHost, cmpopts.IgnoreUnexported(networkingv1beta1api.ServiceEntry{}, networkingv1beta1api.ServicePort{})))
		})
	})
})
