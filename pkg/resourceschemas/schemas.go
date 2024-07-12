package resourceschemas

import (
	"fmt"
	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	nais_io_v1 "github.com/nais/liberator/pkg/apis/nais.io/v1"
	pov1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	policyv1 "k8s.io/api/policy/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

func addGVK(lists []client.ObjectList, scheme *runtime.Scheme) []unstructured.UnstructuredList {
	listsWithGVKs := make([]unstructured.UnstructuredList, 0)
	for _, list := range lists {
		unstructuredList := unstructured.UnstructuredList{}
		gvk, err := apiutil.GVKForObject(list, scheme)
		if err != nil {
			panic(fmt.Errorf("failed to get GVK for object, cant start without schemas: %w", err))
		}
		unstructuredList.SetGroupVersionKind(gvk)
		listsWithGVKs = append(listsWithGVKs, unstructuredList)
	}
	return listsWithGVKs
}

func GetApplicationSchemas(scheme *runtime.Scheme) []unstructured.UnstructuredList {
	return addGVK([]client.ObjectList{
		&appsv1.DeploymentList{},
		&corev1.ServiceList{},
		&corev1.ConfigMapList{},
		&networkingv1beta1.ServiceEntryList{},
		&networkingv1beta1.GatewayList{},
		&autoscalingv2.HorizontalPodAutoscalerList{},
		&networkingv1beta1.VirtualServiceList{},
		&securityv1beta1.PeerAuthenticationList{},
		&corev1.ServiceAccountList{},
		&policyv1.PodDisruptionBudgetList{},
		&networkingv1.NetworkPolicyList{},
		&securityv1beta1.AuthorizationPolicyList{},
		&nais_io_v1.MaskinportenClientList{},
		&nais_io_v1.IDPortenClientList{},
		&pov1.ServiceMonitorList{},
		&certmanagerv1.CertificateList{},
	}, scheme)
}
