package requestauthentication

import (
	"fmt"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/api/v1alpha1/istiotypes"
	"github.com/kartverket/skiperator/pkg/reconciliation"
	"github.com/kartverket/skiperator/pkg/util"
	securityv1api "istio.io/api/security/v1"
	"istio.io/api/security/v1beta1"
	typev1beta1 "istio.io/api/type/v1beta1"
	securityv1 "istio.io/client-go/pkg/apis/security/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Generate(r reconciliation.Reconciliation) error {
	ctxLog := r.GetLogger()
	if r.GetType() != reconciliation.ApplicationType {
		return fmt.Errorf("unsupported type %s in RequestAuthentication", r.GetType())
	}
	application, ok := r.GetSKIPObject().(*skiperatorv1alpha1.Application)
	if !ok {
		err := fmt.Errorf("failed to cast resource to application")
		ctxLog.Error(err, "Failed to generate RequestAuthentication")
		return err
	}
	ctxLog.Debug("Attempting to generate RequestAuthentication for application", "application", application.Name)

	providers := getIdentityProviders(r, application)
	if provider != nil {
	}

	requestAuthentication := getRequestAuthentication(application, providers)

	ctxLog.Debug("Finished generating RequestAuthentication for application", "application", application.Name)
	r.AddResource(&requestAuthentication)

	return nil
}

func getIdentityProviders(r reconciliation.Reconciliation, application *skiperatorv1alpha1.Application) *[]IdentityProvider {
	ctxLog := r.GetLogger()
	switch {
	case application.Spec.IDPorten != nil && application.Spec.Maskinporten == nil:
		return &[]IdentityProvider{
			MASKINPORTEN,
		}
	case application.Spec.IDPorten == nil && application.Spec.Maskinporten != nil:
		return &[]IdentityProvider{
			ID_PORTEN,
		}
	case application.Spec.IDPorten == nil && application.Spec.Maskinporten == nil:
		return &[]IdentityProvider{
			MASKINPORTEN, ID_PORTEN,
		}
	default:
		ctxLog.Debug("No RequestAuthentication will be generated without the presence of neither IDPorten nor Maskinporten.")
		return nil
	}
}

func getRequestAuthentication(application *skiperatorv1alpha1.Application, providers []IdentityProvider) securityv1.RequestAuthentication {
	jwtRules := make([]v1beta1.JWTRule, len(providers))
	for i, provider := range providers {
		if provider == ID_PORTEN {
			jwtRules[i] = getJWTRule(application.Spec.IDPorten.Authentication, provider)
		} else if provider == MASKINPORTEN {
			jwtRules[i] = getJWTRule(application.Spec.Maskinporten.Authentication, provider)
		}
	}

	return securityv1.RequestAuthentication{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: application.Namespace,
			Name:      application.Name + "-jwt-authn",
		},
		Spec: securityv1api.RequestAuthentication{
			Selector: &typev1beta1.WorkloadSelector{
				MatchLabels: util.GetPodAppSelector(application.Name),
			},
			JwtRules:,
		},
	}
}

func getJWTRule(authentication *istiotypes.Authentication, provider IdentityProvider) v1beta1.JWTRule {
	var forwardOriginalToken = true
	if authentication.ForwardOriginalToken != nil {
		forwardOriginalToken = *authentication.ForwardOriginalToken
	}
	var requestAuthentication = v1beta1.JWTRule{
		Audiences:            []string{}, //TODO: Retrieve audience
		ForwardOriginalToken: forwardOriginalToken,
	}
	if authentication.TokenLocation == &HEADER {
		requestAuthentication.FromHeaders = []*v1beta1.JWTHeader{
			{
				Name:   "Authorization",
				Prefix: "Bearer ",
			},
		}
	} else if authentication.TokenLocation == &COOKIE {
		requestAuthentication.FromCookies = []string{"BearerToken"}
	}

	switch provider {
	case ID_PORTEN:
		requestAuthentication.Issuer = ""
		requestAuthentication.JwksUri = ""
		if authentication.TokenLocation == nil {
			requestAuthentication.FromCookies = []string{"BearerToken"}
		}

	case MASKINPORTEN:
		requestAuthentication.Issuer = ""
		requestAuthentication.JwksUri = ""
		if authentication.TokenLocation == nil {
			requestAuthentication.FromHeaders = []*v1beta1.JWTHeader{
				{
					Name:   "Authorization",
					Prefix: "Bearer ",
				},
			}
		}
	}
	return requestAuthentication
}

type IdentityProvider string

var (
	MASKINPORTEN IdentityProvider = "MASKINPORTEN"
	ID_PORTEN    IdentityProvider = "ID_PORTEN"
)

var (
	HEADER = "HEADER"
	COOKIE = "COOKIE"
)
