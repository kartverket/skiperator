package certs

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/kartverket/skiperator/pkg/util"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func AssertWellKnownTlsCert(c client.Client, ctx context.Context, certName string) (bool, []string, error) {
	secret, err := util.GetSecret(c, ctx, types.NamespacedName{Namespace: "istio-gateways", Name: certName})
	if err != nil {
		return false, nil, err
	}

	if secret.Type != corev1.SecretTypeTLS {
		return false, nil, fmt.Errorf("secret %s/%s is not a TLS secret. actual type: %s", secret.Namespace, secret.Name, secret.Type)
	}

	certData := secret.Data[corev1.TLSCertKey]
	if len(certData) == 0 {
		return false, nil, fmt.Errorf("secret %s/%s does not contain a valid certificate", secret.Namespace, secret.Name)
	}

	certKeyData := secret.Data[corev1.TLSPrivateKeyKey]
	if len(certKeyData) == 0 {
		return false, nil, fmt.Errorf("secret %s/%s does not contain a valid private key", secret.Namespace, secret.Name)
	}

	cert, err := tls.X509KeyPair(certData, certKeyData)
	if err != nil {
		return false, nil, fmt.Errorf("failed to parse certificate/key pair: %w", err)
	}

	x509.VerifyOptions{}
}
