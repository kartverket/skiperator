package entraid

import (
	"fmt"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/api/v1alpha1/idprovider"
	"github.com/kartverket/skiperator/pkg/reconciliation"
	"github.com/kartverket/skiperator/pkg/util"
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
		return fmt.Errorf("unsupported type %s in entraid resource", r.GetType())
	}
	application, ok := r.GetSKIPObject().(*skiperatorv1alpha1.Application)
	if !ok {
		err := fmt.Errorf("failed to cast resource to application")
		ctxLog.Error(err, "Failed to generate entraid resource")
		return err
	}
	if application.Spec.EntraID == nil {
		return nil
	}
	if !application.Spec.EntraID.Enabled {
		ctxLog.Debug("Do not create AzureAdApplication resource as Entra ID registration is disabled for application", "application", application.Name)
		return nil
	}

	ctxLog.Debug("Attempting to generate AzureAdApplication resource for application", "application", application.Name)

	azureAdApplicationSpec := naisiov1.AzureAdApplicationSpec{
		AllowAllUsers:             &application.Spec.EntraID.AllowAllUsers,
		PreAuthorizedApplications: application.Spec.EntraID.PreAuthorizedApplications,
	}

	if application.Spec.EntraID.Claims != nil {
		azureAdApplicationSpec.Claims = &naisiov1.AzureAdClaims{
			Groups: application.Spec.EntraID.Claims.Groups,
		}
	}

	if application.Spec.EntraID.LogoutUrl == nil {
		if len(application.Spec.Ingresses) > 0 {
			azureAdApplicationSpec.LogoutUrl = fmt.Sprintf("https://%s%s", application.Spec.Ingresses[0], DefaultClientLogoutPath)
		} else {
			azureAdApplicationSpec.LogoutUrl = fmt.Sprintf("%s%s", KVBaseURL, DefaultClientLogoutPath)
		}
	} else {
		azureAdApplicationSpec.LogoutUrl = *application.Spec.EntraID.LogoutUrl
	}

	if len(application.Spec.EntraID.ReplyUrls) == 0 {
		var replyUrls []naisiov1.AzureAdReplyUrl
		if len(application.Spec.Ingresses) > 0 {
			for _, ingress := range application.Spec.Ingresses {
				replyUrls = append(replyUrls, naisiov1.AzureAdReplyUrl{
					Url: naisiov1.AzureAdReplyUrlString(fmt.Sprintf("https://%s%s", ingress, DefaultClientCallbackPath)),
				})
			}
		}
		azureAdApplicationSpec.ReplyUrls = replyUrls
	} else {
		azureAdApplicationSpec.ReplyUrls = application.Spec.EntraID.ReplyUrls
	}

	secretName, err := GetEntraIdSecretName(application)
	if err != nil {
		ctxLog.Error(err, "Failed to generate secret name for AzureAdApplication resource for application", "application", application.Name)
		return err
	}
	azureAdApplicationSpec.SecretName = secretName

	if application.Spec.EntraID.SecretKeyPrefix != nil {
		azureAdApplicationSpec.SecretKeyPrefix = *application.Spec.EntraID.SecretKeyPrefix
	}

	if application.Spec.EntraID.SecretProtected != nil {
		azureAdApplicationSpec.SecretProtected = *application.Spec.EntraID.SecretProtected
	}

	if application.Spec.EntraID.SinglePageApplication != nil {
		azureAdApplicationSpec.SinglePageApplication = application.Spec.EntraID.SinglePageApplication
	}

	azureAdApplication := naisiov1.AzureAdApplication{
		ObjectMeta: metav1.ObjectMeta{
			Name:      application.Name,
			Namespace: application.Namespace,
		},
		Spec: azureAdApplicationSpec,
	}

	r.AddResource(&azureAdApplication)
	ctxLog.Debug("Finished generating AzureAdApplication resource for application", "application", application.Name)

	return nil
}

func EntraIDSpecifiedInSpec(entraIDSpec *idprovider.EntraID) bool {
	return entraIDSpec != nil && entraIDSpec.Enabled
}

func GetEntraIdSecretName(application *skiperatorv1alpha1.Application) (string, error) {
	if application.Spec.EntraID.SecretName != nil {
		return *application.Spec.EntraID.SecretName, nil
	} else {
		secretName, err := util.GetSecretName("entraid", application.Name)
		if err != nil {
			return "", err
		}
		return secretName, nil
	}
}
