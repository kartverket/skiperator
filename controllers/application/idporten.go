package applicationcontroller

import (
	"context"
	"net/url"
	"path"

	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/util"
	"github.com/kartverket/skiperator/pkg/util/array"
	nais_io_v1 "github.com/nais/liberator/pkg/apis/nais.io/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	digdiratorTypes "github.com/nais/digdirator/pkg/digdir/types"
)

const (
	DefaultClientCallbackPath = "/oauth2/callback"
	DefaultClientLogoutPath   = "/oauth2/logout"
	DefaultIntegrationType    = string(digdiratorTypes.IntegrationTypeUnknown)

	KVBaseURL = "https://kartverket.no"
)

func (r *ApplicationReconciler) reconcileIDPorten(ctx context.Context, application *skiperatorv1alpha1.Application) (reconcile.Result, error) {
	controllerName := "IDPorten"
	r.SetControllerProgressing(ctx, application, controllerName)

	secretName, err := util.GetSecretName("idporten", application.Name)
	if err != nil {
		r.SetControllerError(ctx, application, controllerName, err)
		return reconcile.Result{}, err
	}

	idporten := nais_io_v1.IDPortenClient{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "nais.io/v1",
			Kind:       "IDPortenClient",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: application.Namespace,
			Name:      application.Name,
		},
	}

	if idportenSpecifiedInSpec(application.Spec.IDPorten) {
		_, err = ctrlutil.CreateOrPatch(ctx, r.GetClient(), &idporten, func() error {
			// Set application as owner of the sidecar
			err := ctrlutil.SetControllerReference(application, &idporten, r.GetScheme())
			if err != nil {
				r.SetControllerError(ctx, application, controllerName, err)
				return err
			}

			r.SetLabelsFromApplication(ctx, &idporten, *application)
			util.SetCommonAnnotations(&idporten)

			ingress := KVBaseURL
			if len(application.Spec.Ingresses) != 0 {
				ingress = application.Spec.Ingresses[0]
			}
			ingress = util.EnsurePrefix(ingress, "https://") // TODO: account for redirectToHTTPS?

			// Use given integration type if set, else use default
			integrationType := WithFallbackStr(application.Spec.IDPorten.IntegrationType, DefaultIntegrationType)

			// Is scopes is set: use that. Else: use defaults
			// var scopes []string
			// if scopes != nil {
			// 	scopes = *application.Spec.IDPorten.Scopes
			// } else {
			// 	scopes = GetDefaultScopesForIntegration(integrationType)
			// }
			scopes := GetScopesForIntegration(application.Spec.IDPorten.IntegrationType, application.Spec.IDPorten.Scopes)

			redirectURIs, err := BuildURIs(application.Spec.Ingresses, application.Spec.IDPorten.RedirectPath, DefaultClientCallbackPath)
			if err != nil {
				r.SetControllerError(ctx, application, controllerName, err)
				return err
			}

			frontchannelLogoutURI, err := BuildURI(ingress, application.Spec.IDPorten.FrontchannelLogoutPath, DefaultClientLogoutPath)
			if err != nil {
				r.SetControllerError(ctx, application, controllerName, err)
				return err
			}

			// prefer given redirectURIs, but default to <ingress>/<DefaultClientLogoutPath> if nothing is specified
			var postLogoutRedirectURIs []nais_io_v1.IDPortenURI
			if application.Spec.IDPorten.PostLogoutRedirectURIs != nil {
				// uri => <uri>/<logout-path>
				postLogoutRedirectURIs, err = array.MapErr(*application.Spec.IDPorten.PostLogoutRedirectURIs, func(uri nais_io_v1.IDPortenURI) (nais_io_v1.IDPortenURI, error) {
					u, err := BuildURI(ingress, "", DefaultClientLogoutPath)
					return nais_io_v1.IDPortenURI(u), err
				})
				if err != nil {
					r.SetControllerError(ctx, application, controllerName, err)
					return err
				}
			} else {
				uri, err := BuildURI(ingress, application.Spec.IDPorten.PostLogoutRedirectPath, DefaultClientLogoutPath)
				if err != nil {
					r.SetControllerError(ctx, application, controllerName, err)
					return err
				}

				postLogoutRedirectURIs = []nais_io_v1.IDPortenURI{nais_io_v1.IDPortenURI(uri)}
			}

			idporten.Spec = nais_io_v1.IDPortenClientSpec{
				ClientURI:              nais_io_v1.IDPortenURI(WithFallbackStr(application.Spec.IDPorten.ClientURI, nais_io_v1.IDPortenURI(ingress))),
				IntegrationType:        integrationType,
				RedirectURIs:           redirectURIs,
				SecretName:             secretName,
				AccessTokenLifetime:    WithFallback(application.Spec.IDPorten.AccessTokenLifetime, 3600),
				SessionLifetime:        WithFallback(application.Spec.IDPorten.SessionLifetime, 7200),
				FrontchannelLogoutURI:  nais_io_v1.IDPortenURI(frontchannelLogoutURI),
				PostLogoutRedirectURIs: postLogoutRedirectURIs,
				Scopes:                 scopes,
			}

			return nil
		})

		if err != nil {
			r.SetControllerError(ctx, application, controllerName, err)
			return reconcile.Result{}, err
		}
	} else {
		err = r.GetClient().Delete(ctx, &idporten)
		err = client.IgnoreNotFound(err)
		if err != nil {
			r.SetControllerError(ctx, application, controllerName, err)
			return reconcile.Result{}, err
		}
	}

	r.SetControllerFinishedOutcome(ctx, application, controllerName, err)

	return reconcile.Result{}, err
}

func GetDefaultScopesForIntegration(integration string) []string {
	switch digdiratorTypes.IntegrationType(integration) {
	case digdiratorTypes.IntegrationTypeKrr:
		return []string{"krr:global/kontaktinformasjon.read", "krr:global/digitalpost.read"}
	case digdiratorTypes.IntegrationTypeIDPorten:
		return []string{"profile", "openid"}
	}

	return make([]string, 0)
}

func GetScopesForIntegration(integration string, scopes *[]string) []string {
	scps := make([]string, 0)

	if defaultScopes := GetDefaultScopesForIntegration(integration); len(defaultScopes) != 0 {
		scps = defaultScopes
	} else if scopes != nil {
		scps = *scopes
	}

	return scps
}

func WithFallback[T any](val *T, fallback T) *T {
	if val == nil {
		return &fallback
	}

	return val
}

func WithFallbackStr[T ~string](val T, fallback T) T {
	if val == "" {
		return fallback
	}

	return val
}

// https://github.com/nais/naiserator/blob/423676cb2415cfd2ca40ba8e6e5d9edb46f15976/pkg/util/url.go#L31-L35
func AppendToPath(ingress string, pathSeg string) (string, error) {
	u, err := url.Parse(ingress)
	if err != nil {
		return "", err // parse errors have very understandable error message by default -> no need to wrap
	}

	u.Path = path.Join(u.Path, pathSeg)
	return u.String(), nil
}

func BuildURI(ingress string, pathSeg string, fallback string) (string, error) {
	if pathSeg == "" {
		pathSeg = fallback
	}

	ingress = util.EnsurePrefix(ingress, "https://")

	return AppendToPath(ingress, pathSeg)
}

func BuildURIs(ingresses []string, pathSeg string, fallback string) ([]nais_io_v1.IDPortenURI, error) {
	return array.MapErr(ingresses, func(ingress string) (nais_io_v1.IDPortenURI, error) {
		uri, err := BuildURI(ingress, pathSeg, fallback)
		return nais_io_v1.IDPortenURI(uri), err
	})
}

func idportenSpecifiedInSpec(mp *skiperatorv1alpha1.IDPorten) bool {
	return mp != nil && mp.Enabled
}
