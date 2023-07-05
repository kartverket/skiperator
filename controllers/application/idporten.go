package applicationcontroller

import (
	"context"
	"fmt"
	"net/url"
	"path"

	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/util"
	"github.com/kartverket/skiperator/pkg/util/array"
	naisv1 "github.com/nais/liberator/pkg/apis/nais.io/v1"
	"github.com/nais/liberator/pkg/namegen"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	DefaultClientCallbackPath = "/oauth2/callback"
	DefaultClientLogoutPath   = "/oauth2/logout"
)

type DigdiratorConnectionConfig struct {
	// path to kubeconfig file. e.g. "$HOME/.kube/config"
	kubeconfig string
	// URL of the server
	masterURL string
	// namespace of digdirator
	namespace string
}

func (r *ApplicationReconciler) reconcileIDPorten(ctx context.Context, application *skiperatorv1alpha1.Application) (reconcile.Result, error) {
	controllerName := "IDPorten"
	r.SetControllerProgressing(ctx, application, controllerName)

	// // TODO: add conn conf to spec?
	// connConfig := DigdiratorConnectionConfig{}

	// digdiratorClient, err := connConfig.GetClient()
	// if err != nil {
	// 	return reconcile.Result{}, fmt.Errorf("failed to initialize digdirator client %w", err)
	// }

	if application.Spec.IDPorten == nil || !application.Spec.IDPorten.Enabled {
		return reconcile.Result{}, fmt.Errorf("idporten is not enabled for this application")
	}

	// https://github.com/nais/naiserator/blob/faed273b68dff8541e1e2889fda5d017730f9796/pkg/resourcecreator/idporten/idporten.go#L82
	// https://github.com/nais/naiserator/blob/faed273b68dff8541e1e2889fda5d017730f9796/pkg/resourcecreator/idporten/idporten.go#L170
	secretName, err := namegen.ShortName(fmt.Sprintf("idporten-%s/%s", application.Namespace, application.Name), validation.DNS1035LabelMaxLength)
	if err != nil {
		return reconcile.Result{}, err
	}

	var idporten naisv1.IDPortenClient

	if err := r.GetClient().Get(ctx, types.NamespacedName{
		Namespace: idporten.ObjectMeta.Namespace,
		Name:      idporten.ObjectMeta.Name,
	}, &idporten); err != nil {
		return reconcile.Result{}, err
	}

	_, err = ctrlutil.CreateOrPatch(ctx, r.GetClient(), &idporten, func() error {
		// Set application as owner of the sidecar
		err := ctrlutil.SetControllerReference(application, &idporten, r.GetScheme())
		if err != nil {
			r.SetControllerError(ctx, application, controllerName, err)
			return err
		}

		r.SetLabelsFromApplication(ctx, &idporten, *application)
		util.SetCommonAnnotations(&idporten)

		redirectURIs, err := BuildURIs(application.Spec.Ingresses, application.Spec.IDPorten.RedirectPath, DefaultClientCallbackPath)
		if err != nil {
			r.SetControllerError(ctx, application, controllerName, err)
			return err
		}

		frontchannelLogoutURI, err := BuildURI(application.Spec.Ingresses[0], application.Spec.IDPorten.FrontchannelLogoutPath, DefaultClientLogoutPath)
		if err != nil {
			r.SetControllerError(ctx, application, controllerName, err)
			return err
		}

		spec := naisv1.IDPortenClientSpec{
			ClientURI:       application.Spec.IDPorten.ClientURI,
			IntegrationType: application.Spec.IDPorten.IntegrationType,
			RedirectURIs:    redirectURIs,
			SecretName:      secretName,
			// note fallback might not be needed (see: AccessTokenLifetime and SessionLifetime docs, has defaults)
			AccessTokenLifetime:    WithFallback(application.Spec.IDPorten.AccessTokenLifetime, 3600),
			SessionLifetime:        WithFallback(application.Spec.IDPorten.SessionLifetime, 7200),
			FrontchannelLogoutURI:  naisv1.IDPortenURI(frontchannelLogoutURI),
			PostLogoutRedirectURIs: []naisv1.IDPortenURI{application.Spec.IDPorten.ClientURI},
		}

		idporten.Spec = spec

		return nil
	})

	r.SetControllerFinishedOutcome(ctx, application, controllerName, err)

	return reconcile.Result{}, err
}

func WithFallback[T any](val *T, fallback T) *T {
	if val == nil {
		val = &fallback
	}

	return val
}

func Coalesce[T comparable](vals ...T) *T {
	var null T

	for _, v := range vals {
		if v != null {
			return &v
		}
	}

	return nil
}

func Must[T any](d T, err error) T {
	if err != nil {
		panic(err)
	}

	return d
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

	return AppendToPath(ingress, pathSeg)
}

func BuildURIs(ingresses []string, pathSeg string, fallback string) ([]naisv1.IDPortenURI, error) {
	return array.MapErr(ingresses, func(ingress string) (naisv1.IDPortenURI, error) {
		uri, err := BuildURI(ingress, pathSeg, fallback)
		return naisv1.IDPortenURI(uri), err
	})
}

// Does not work
func (dcc *DigdiratorConnectionConfig) GetClient() (*kubernetes.Clientset, error) {
	conf, err := clientcmd.BuildConfigFromFlags(dcc.masterURL, dcc.kubeconfig)
	if err != nil {
		return nil, err
	}

	client, err := kubernetes.NewForConfig(conf)

	return client, err
}
