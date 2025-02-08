package maskinporten

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

type MaskinportenOps struct{}

func (i *MaskinportenOps) Provider() jwtAuth.IdentityProvider {
	return jwtAuth.MASKINPORTEN
}

func (i *MaskinportenOps) IssuerKey() string {
	return secrets.MaskinportenIssuerKey
}

func (i *MaskinportenOps) JwksKey() string {
	return secrets.MaskinportenJwksUriKey
}

func (i *MaskinportenOps) ClientIDKey() string {
	return secrets.MaskinportenClientIDKey
}

func (i *MaskinportenOps) GetSecretName(k8sClient client.Client, ctx context.Context, application skiperatorv1alpha1.Application) (*string, error) {
	namespacedName := types.NamespacedName{
		Namespace: application.Namespace,
		Name:      application.Name,
	}
	maskinportenClient, err := util.GetMaskinPortenlient(k8sClient, ctx, namespacedName)
	if err != nil {
		return nil, fmt.Errorf("failed to get MaskinportenClient: %s", namespacedName.String())
	}
	if maskinportenClient == nil {
		return nil, fmt.Errorf("MaskinportenClient: '%s' not found", namespacedName.String())
	}
	for _, ownerReference := range maskinportenClient.OwnerReferences {
		if ownerReference.UID == application.UID {
			return &maskinportenClient.Spec.SecretName, nil
		}
	}
	return nil, fmt.Errorf("no MaskinportenClient with ownerRef to '%s' found", namespacedName.String())
}

func (i *MaskinportenOps) GetSecretData(k8sClient client.Client, ctx context.Context, namespace, secretName string) (map[string][]byte, error) {
	return util.GetSecretData(k8sClient, ctx, types.NamespacedName{Namespace: namespace, Name: secretName},
		[]string{i.IssuerKey(), i.JwksKey(), i.ClientIDKey()})
}

func (i *MaskinportenOps) GetProviderInfo(k8sClient client.Client, ctx context.Context, application skiperatorv1alpha1.Application) (*jwtAuth.IdentityProviderInfo, error) {
	var secretName *string
	var err error
	if application.Spec.Maskinporten.Authentication.SecretName != nil {
		secretName = application.Spec.Maskinporten.Authentication.SecretName
	} else if application.Spec.Maskinporten.Enabled {
		secretName, err = i.GetSecretName(k8sClient, ctx, application)
	} else {
		return nil, fmt.Errorf("JWT authentication requires either Maskinporten to be enabled or a secretName to be provided")
	}
	if err != nil {
		err := fmt.Errorf("failed to get secret name for MaskinportenClient: %w", err)
		return nil, err
	}

	var notPaths *[]string
	if application.Spec.Maskinporten.Authentication.IgnorePaths != nil {
		notPaths = application.Spec.Maskinporten.Authentication.IgnorePaths
	} else {
		notPaths = nil
	}

	return &jwtAuth.IdentityProviderInfo{
		Provider:   jwtAuth.MASKINPORTEN,
		SecretName: *secretName,
		NotPaths:   notPaths,
	}, nil
}
