package idporten

import (
	"fmt"
	"net/url"
	"path"

	skiperatorv1beta1 "github.com/kartverket/skiperator/api/v1beta1"
	"github.com/kartverket/skiperator/api/v1beta1/digdirator"
	"github.com/kartverket/skiperator/pkg/reconciliation"
	"github.com/kartverket/skiperator/pkg/util"
	"github.com/kartverket/skiperator/pkg/util/array"
	digdiratorClients "github.com/nais/digdirator/pkg/clients"
	digdiratorTypes "github.com/nais/digdirator/pkg/digdir/types"
	naisiov1 "github.com/nais/liberator/pkg/apis/nais.io/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	DefaultClientCallbackPath = "/oauth2/callback"
	DefaultClientLogoutPath   = "/oauth2/logout"

	KVBaseURL = "https://kartverket.no"
)

func Generate(r reconciliation.Reconciliation) error {
	ctxLog := r.GetLogger()
	if r.GetType() != reconciliation.ApplicationType {
		return fmt.Errorf("unsupported type %s in idporten resource", r.GetType())
	}
	application, ok := r.GetSKIPObject().(*skiperatorv1beta1.Application)
	if !ok {
		err := fmt.Errorf("failed to cast resource to application")
		ctxLog.Error(err, "Failed to generate idporten resource")
		return err
	}
	if application.Spec.IDPorten == nil {
		return nil
	}
	ctxLog.Debug("Attempting to generate id porten resource for application", "application", application.Name)

	var err error

	idporten := naisiov1.IDPortenClient{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "nais.io/v1",
			Kind:       "IDPortenClient",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: application.Namespace,
			Name:      application.Name,
		},
	}

	idporten.Spec, err = getIDPortenSpec(application)
	if err != nil {
		return err
	}

	r.AddResource(&idporten)
	ctxLog.Debug("Finished generating id porten resource for application", "application", application.Name)

	return nil
}

// Assumes application.Spec.IDPorten != nil
func getIDPortenSpec(application *skiperatorv1beta1.Application) (naisiov1.IDPortenClientSpec, error) {
	integrationType := application.Spec.IDPorten.IntegrationType
	if integrationType == "" {
		// No scopes => idporten
		// Scopes    => api_klient
		if len(application.Spec.IDPorten.Scopes) == 0 {
			integrationType = string(digdiratorTypes.IntegrationTypeIDPorten)
		} else {
			integrationType = string(digdiratorTypes.IntegrationTypeApiKlient)
		}
	}

	ingress := KVBaseURL
	hosts, err := application.Spec.Hosts()
	if err != nil {
		return naisiov1.IDPortenClientSpec{}, err
	}

	ingresses := hosts.Hostnames()
	if len(ingresses) != 0 {
		ingress = ingresses[0]
	}
	ingress = util.EnsurePrefix(ingress, "https://")

	scopes := getScopes(integrationType, application.Spec.IDPorten.Scopes)

	redirectURIs, err := buildURIs(ingresses, application.Spec.IDPorten.RedirectPath, DefaultClientCallbackPath)
	if err != nil {
		return naisiov1.IDPortenClientSpec{}, nil
	}

	frontchannelLogoutURI, err := buildURI(ingress, application.Spec.IDPorten.FrontchannelLogoutPath, DefaultClientLogoutPath)
	if err != nil {
		return naisiov1.IDPortenClientSpec{}, nil
	}

	postLogoutRedirectURIs, err := getPostLogoutRedirectURIs(application.Spec.IDPorten.PostLogoutRedirectURIs, ingress, application.Spec.IDPorten.PostLogoutRedirectPath)
	if err != nil {
		return naisiov1.IDPortenClientSpec{}, nil
	}

	secretName, err := GetIDPortenSecretName(application.Name)
	if err != nil {
		return naisiov1.IDPortenClientSpec{}, err
	}

	return naisiov1.IDPortenClientSpec{
		ClientName:             getClientNameIdPorten(application.Name, application.Spec.IDPorten),
		ClientURI:              withFallback(application.Spec.IDPorten.ClientURI, naisiov1.IDPortenURI(ingress)),
		IntegrationType:        integrationType,
		RedirectURIs:           redirectURIs,
		SecretName:             secretName,
		AccessTokenLifetime:    application.Spec.IDPorten.AccessTokenLifetime,
		SessionLifetime:        application.Spec.IDPorten.SessionLifetime,
		FrontchannelLogoutURI:  naisiov1.IDPortenURI(frontchannelLogoutURI),
		PostLogoutRedirectURIs: postLogoutRedirectURIs,
		Scopes:                 scopes,
	}, nil
}

func getClientNameIdPorten(applicationName string, idPortenSettings *digdirator.IDPorten) string {
	if idPortenSettings.ClientName != nil {
		return *idPortenSettings.ClientName
	}

	return applicationName
}

func getPostLogoutRedirectURIs(postLogoutRedirectURIs *[]naisiov1.IDPortenURI, ingress string, postLogoutRedirectPath string) ([]naisiov1.IDPortenURI, error) {
	uris := make([]naisiov1.IDPortenURI, 0)

	if postLogoutRedirectURIs != nil {
		uris = *postLogoutRedirectURIs
	}

	if postLogoutRedirectPath != "" {
		u, err := buildURI(ingress, postLogoutRedirectPath, DefaultClientLogoutPath)
		if err != nil {
			return uris, err
		}
		uris = append(uris, naisiov1.IDPortenURI(u))
	}

	return uris, nil
}

func getScopes(integrationType string, scopes []string) []string {
	defaultScopes := digdiratorClients.GetIDPortenDefaultScopes(integrationType)
	if len(defaultScopes) != 0 {
		return defaultScopes
	}

	return scopes
}

func withFallback[T ~string](val T, fallback T) T {
	if val == "" {
		return fallback
	}

	return val
}

// https://github.com/nais/naiserator/blob/423676cb2415cfd2ca40ba8e6e5d9edb46f15976/pkg/util/url.go#L31-L35
func appendToPath(ingress string, pathSeg string) (string, error) {
	u, err := url.Parse(ingress)
	if err != nil {
		return "", err
	}

	u.Path = path.Join(u.Path, pathSeg)
	return u.String(), nil
}

// Ensures "https://" prefix and adds path if given, if not adds fallback
// uri, path, fallback => https://<uri>/<path>
// uri, "", fallback   => https://<uri>/<fallback>
func buildURI(ingress string, pathSeg string, fallback string) (string, error) {
	if pathSeg == "" {
		pathSeg = fallback
	}

	ingress = util.EnsurePrefix(ingress, "https://")

	return appendToPath(ingress, pathSeg)
}

// ingress => BuildURI(ingress)
func buildURIs(ingresses []string, pathSeg string, fallback string) ([]naisiov1.IDPortenURI, error) {
	return array.MapErr(ingresses, func(ingress string) (naisiov1.IDPortenURI, error) {
		uri, err := buildURI(ingress, pathSeg, fallback)
		return naisiov1.IDPortenURI(uri), err
	})
}

func IdportenSpecifiedInSpec(mp *digdirator.IDPorten) bool {
	return mp != nil && mp.Enabled
}

func GetIDPortenSecretName(name string) (string, error) {
	return util.GetSecretName("idporten", name)
}
