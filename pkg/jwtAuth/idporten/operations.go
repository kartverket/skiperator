package idporten

import (
	"context"
	"fmt"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/jwtAuth"
	"github.com/kartverket/skiperator/pkg/util"
	"github.com/nais/digdirator/pkg/secrets"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type IDPortenOps struct{}

func (i *IDPortenOps) Provider() jwtAuth.IdentityProvider {
	return jwtAuth.ID_PORTEN
}

func (i *IDPortenOps) IssuerKey() string {
	return secrets.IDPortenIssuerKey
}

func (i *IDPortenOps) JwksKey() string {
	return secrets.IDPortenJwksUriKey
}

func (i *IDPortenOps) ClientIDKey() string {
	return secrets.IDPortenClientIDKey
}

func (i *IDPortenOps) GetSecretName(k8sClient client.Client, ctx context.Context, application skiperatorv1alpha1.Application) (*string, error) {
	namespacedName := types.NamespacedName{
		Namespace: application.Namespace,
		Name:      application.Name,
	}
	idPortenClient, err := util.GetIdPortenClient(k8sClient, ctx, namespacedName)
	if err != nil {
		return nil, fmt.Errorf("failed to get IDPortenClient: %s", namespacedName.String())
	}
	if idPortenClient == nil {
		return nil, fmt.Errorf("IDPortenClient: '%s' not found", namespacedName.String())
	}
	for _, ownerReference := range idPortenClient.OwnerReferences {
		if ownerReference.UID == application.UID {
			return &idPortenClient.Spec.SecretName, nil
		}
	}
	return nil, fmt.Errorf("no IDPortenClient with ownerRef to '%s' found", namespacedName.String())
}

func (i *IDPortenOps) GetSecretData(k8sClient client.Client, ctx context.Context, namespace, secretName string) (map[string][]byte, error) {
	return util.GetSecretData(k8sClient, ctx, types.NamespacedName{Namespace: namespace, Name: secretName},
		[]string{i.IssuerKey(), i.JwksKey(), i.ClientIDKey()})
}

func (i *IDPortenOps) GetProviderInfo(k8sClient client.Client, ctx context.Context, application skiperatorv1alpha1.Application) (*jwtAuth.IdentityProviderInfo, error) {
	var secretName *string
	var err error
	if application.Spec.IDPorten.Authentication.SecretName != nil {
		secretName = application.Spec.IDPorten.Authentication.SecretName
	} else if application.Spec.IDPorten.Enabled {
		secretName, err = i.GetSecretName(k8sClient, ctx, application)
	} else {
		return nil, fmt.Errorf("JWT authentication requires either IDPorten to be enabled or a secretName to be provided")
	}
	if err != nil {
		err := fmt.Errorf("failed to get secret name for IDPortenClient: %w", err)
		return nil, err
	}

	var notPaths *[]string
	if application.Spec.IDPorten.Authentication.IgnorePaths != nil {
		notPaths = application.Spec.IDPorten.Authentication.IgnorePaths
	} else {
		notPaths = nil
	}

	return &jwtAuth.IdentityProviderInfo{
		Provider:   jwtAuth.ID_PORTEN,
		SecretName: *secretName,
		NotPaths:   notPaths,
	}, nil
}
