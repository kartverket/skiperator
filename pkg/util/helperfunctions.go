package util

import (
	"context"
	"fmt"
	"github.com/kartverket/skiperator/api/v1alpha1/podtypes"
	"github.com/mitchellh/hashstructure/v2"
	"github.com/nais/liberator/pkg/namegen"
	"hash/fnv"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

// TODO Clean up this file, move functions to more appropriate files

func GetHashForStructs(obj []interface{}) string {
	hash, err := hashstructure.Hash(obj, hashstructure.FormatV2, nil)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%d", hash)
}

func GenerateHashFromName(name string) uint64 {
	hash := fnv.New64()
	_, _ = hash.Write([]byte(name))
	return hash.Sum64()
}

func GetConfigMap(client client.Client, ctx context.Context, namespacedName types.NamespacedName) (corev1.ConfigMap, error) {
	configMap := corev1.ConfigMap{}

	err := client.Get(ctx, namespacedName, &configMap)

	return configMap, err
}

func ErrDoPanic(err error, message string) {
	if err != nil {
		errorMessage := fmt.Sprintf(message, err.Error())
		panic(errorMessage)
	}
}

func PointTo[T any](x T) *T {
	return &x
}

func PointToInt64(n int64) *int64 {
	return &n
}

func GetIstioGatewaySelector() map[string]string {
	return map[string]string{"kubernetes.io/metadata.name": "istio-gateways"}
}

func GetPodAppSelector(applicationName string) map[string]string {
	return map[string]string{"app": applicationName}
}

func GetPodAppAndTeamSelector(applicationName string, teamName string) map[string]string {
	return map[string]string{
		"app":  applicationName,
		"team": teamName,
	}
}

func ResourceNameWithKindPostfix(resourceName string, kind string) string {
	return strings.ToLower(fmt.Sprintf("%v-%v", resourceName, kind))
}

func GetSecretName(prefix string, name string) (string, error) {
	// https://github.com/nais/naiserator/blob/faed273b68dff8541e1e2889fda5d017730f9796/pkg/resourcecreator/idporten/idporten.go#L82
	// https://github.com/nais/naiserator/blob/faed273b68dff8541e1e2889fda5d017730f9796/pkg/resourcecreator/idporten/idporten.go#L170
	secretName, err := namegen.ShortName(fmt.Sprintf("%s-%s", prefix, name), validation.DNS1035LabelMaxLength)
	return secretName, err
}

func EnsurePrefix(s string, prefix string) string {
	if !strings.HasPrefix(s, prefix) {
		return prefix + s
	}
	return s
}

func IsCloudSqlProxyEnabled(gcp *podtypes.GCP) bool {
	return gcp != nil && gcp.CloudSQLProxy != nil
}

func IsGCPAuthEnabled(gcp *podtypes.GCP) bool {
	return gcp != nil && gcp.Auth != nil && gcp.Auth.ServiceAccount != ""
}
