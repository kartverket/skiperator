package authorizationpolicy

import (
	"github.com/kartverket/skiperator/pkg/util"
	v1 "istio.io/api/security/v1"
	"istio.io/api/security/v1beta1"
	typev1beta1 "istio.io/api/type/v1beta1"
	securityv1 "istio.io/client-go/pkg/apis/security/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

const (
	DefaultDenyPath = "/actuator*"
)

func GetAuthPolicy(namespacedName types.NamespacedName, applicationName string, action v1beta1.AuthorizationPolicy_Action, paths []string, notPaths []string) *securityv1.AuthorizationPolicy {
	return &securityv1.AuthorizationPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespacedName.Namespace,
			Name:      namespacedName.Name,
		},
		Spec: v1.AuthorizationPolicy{
			Action: action,
			Rules: []*v1.Rule{
				{
					To: []*v1.Rule_To{
						{
							Operation: &v1.Operation{
								Paths:    paths,
								NotPaths: notPaths,
							},
						},
					},
					From: GetGeneralFromRule(),
				},
			},
			Selector: &typev1beta1.WorkloadSelector{
				MatchLabels: util.GetPodAppSelector(applicationName),
			},
		},
	}
}

func GetGeneralFromRule() []*v1.Rule_From {
	return []*v1.Rule_From{
		{
			Source: &v1.Source{
				Namespaces: []string{"istio-gateways"},
			},
		},
	}
}
