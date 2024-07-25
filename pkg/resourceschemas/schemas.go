package resourceschemas

import (
	"fmt"
	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	nais_io_v1 "github.com/nais/liberator/pkg/apis/nais.io/v1"
	pov1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	policyv1 "k8s.io/api/policy/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	goclientscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

func AddSchemas(scheme *runtime.Scheme) {
	utilruntime.Must(goclientscheme.AddToScheme(scheme))
	utilruntime.Must(skiperatorv1alpha1.AddToScheme(scheme))
	utilruntime.Must(autoscalingv2.AddToScheme(scheme))
	utilruntime.Must(securityv1beta1.AddToScheme(scheme))
	utilruntime.Must(networkingv1beta1.AddToScheme(scheme))
	utilruntime.Must(certmanagerv1.AddToScheme(scheme))
	utilruntime.Must(policyv1.AddToScheme(scheme))
	utilruntime.Must(pov1.AddToScheme(scheme))
	utilruntime.Must(nais_io_v1.AddToScheme(scheme))
	utilruntime.Must(appsv1.AddToScheme(scheme))
	utilruntime.Must(corev1.AddToScheme(scheme))
}

func addGVKToList(lists []client.ObjectList, scheme *runtime.Scheme) []unstructured.UnstructuredList {
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
	return addGVKToList([]client.ObjectList{
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
		&pov1.PodMonitorList{},
		&certmanagerv1.CertificateList{},
	}, scheme)
}

func GetJobSchemas(scheme *runtime.Scheme) []unstructured.UnstructuredList {
	return addGVKToList([]client.ObjectList{
		&batchv1.CronJobList{},
		&batchv1.JobList{},
		&networkingv1.NetworkPolicyList{},
		&corev1.ServiceAccountList{},
		&networkingv1beta1.ServiceEntryList{},
		&corev1.ConfigMapList{},
		&pov1.PodMonitorList{},
	}, scheme)
}

func GetRoutingSchemas(scheme *runtime.Scheme) []unstructured.UnstructuredList {
	return addGVKToList([]client.ObjectList{
		&certmanagerv1.CertificateList{},
		&networkingv1beta1.GatewayList{},
		&networkingv1.NetworkPolicyList{},
		&networkingv1beta1.VirtualServiceList{},
	}, scheme)
}

func GetNamespaceSchemas(scheme *runtime.Scheme) []unstructured.UnstructuredList {
	return addGVKToList([]client.ObjectList{
		&corev1.NamespaceList{},
		&corev1.ConfigMapList{},
		&networkingv1.NetworkPolicyList{},
		&networkingv1beta1.SidecarList{},
		&corev1.SecretList{},
	}, scheme)
}
