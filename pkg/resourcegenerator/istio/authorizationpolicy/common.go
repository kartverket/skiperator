package authorizationpolicy

import (
	"github.com/kartverket/skiperator/api/v1alpha1/istiotypes"
	"github.com/kartverket/skiperator/pkg/auth"
	v1 "istio.io/api/security/v1"
	"istio.io/api/security/v1beta1"
)

const (
	DefaultDenyPath = "/actuator*"
)

func GetGeneralFromRule() []*v1.Rule_From {
	return []*v1.Rule_From{
		{
			Source: &v1.Source{
				Namespaces: []string{"istio-gateways"},
			},
		},
	}
}

func GetApiSurfaceDiffAsRuleToList(requestMatchers, otherRequestMatchers istiotypes.RequestMatchers) []*v1beta1.Rule_To {
	var diff []*v1beta1.Rule_To
	for _, requestMatcher := range requestMatchers {
		ruleTo := &v1beta1.Rule_To{
			Operation: &v1beta1.Operation{
				Paths:   requestMatcher.Paths,
				Methods: requestMatcher.Methods,
			},
		}
		for _, otherRequestMatcher := range otherRequestMatchers {
			ruleTo.Operation.NotPaths = append(ruleTo.Operation.NotPaths, otherRequestMatcher.Paths...)
		}
		diff = append(diff, ruleTo)
	}
	for _, otherRequestMatcher := range otherRequestMatchers {
		notMethods := otherRequestMatcher.Methods
		if len(notMethods) == 0 {
			notMethods = append(notMethods, auth.AcceptedHttpMethods...)
		}
		diff = append(diff, &v1beta1.Rule_To{
			Operation: &v1beta1.Operation{
				Paths:      otherRequestMatcher.Paths,
				NotMethods: notMethods,
			},
		})
	}
	return diff
}
