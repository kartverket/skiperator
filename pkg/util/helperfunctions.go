package util

import (
	"context"
	"fmt"
	"github.com/kartverket/skiperator/api/v1alpha1/podtypes"
	"github.com/mitchellh/hashstructure/v2"
	"github.com/nais/liberator/pkg/namegen"
	"github.com/r3labs/diff/v3"
	"hash/fnv"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/client-go/tools/record"
	"reflect"
	"regexp"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
	"unicode"
)

//TODO Clean up this file, move functions to more appropriate files

var internalPattern = regexp.MustCompile(`[^.]\.skip\.statkart\.no|[^.]\.kartverket-intern.cloud`)

func IsInternal(hostname string) bool {
	return internalPattern.MatchString(hostname)
}

func GetIstioGatewayLabelSelector(hostname string) map[string]string {
	if IsInternal(hostname) {
		return map[string]string{"app": "istio-ingress-internal"}
	}
	return map[string]string{"app": "istio-ingress-external"}

}

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

func GetSecret(client client.Client, ctx context.Context, namespacedName types.NamespacedName) (corev1.Secret, error) {
	secret := corev1.Secret{}

	err := client.Get(ctx, namespacedName, &secret)

	return secret, err
}

func GetService(client client.Client, ctx context.Context, namespacedName types.NamespacedName) (corev1.Service, error) {
	service := corev1.Service{}

	err := client.Get(ctx, namespacedName, &service)

	return service, err
}

func ErrIsMissingOrNil(recorder record.EventRecorder, err error, message string, object runtime.Object) bool {
	if errors.IsNotFound(err) {
		recorder.Eventf(
			object,
			corev1.EventTypeWarning, "Missing",
			message,
		)
	} else if err != nil {
		return false
	}
	return true
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

func HasUpperCaseLetter(word string) bool {
	for _, letter := range word {
		if unicode.IsUpper(letter) {
			return true
		}
	}

	return false
}

func ResourceNameWithKindPostfix(resourceName string, kind string) string {
	return strings.ToLower(fmt.Sprintf("%v-%v", resourceName, kind))
}

func GetGatewaySecretName(namespace string, name string) string {
	return fmt.Sprintf("%s-%s-ingress", namespace, name)
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

func GetObjectDiff[T any](a T, b T) (diff.Changelog, error) {
	aKind := reflect.ValueOf(a).Kind()
	bKind := reflect.ValueOf(b).Kind()
	if aKind != bKind {
		return nil, fmt.Errorf("The objects to compare are not the same, found obj1: %v, obj2: %v\n", aKind, bKind)
	}
	changelog, err := diff.Diff(a, b)

	if len(changelog) == 0 {
		return nil, err
	}

	return changelog, nil
}

func IsCloudSqlProxyEnabled(gcp *podtypes.GCP) bool {
	return gcp != nil && gcp.CloudSQLProxy.Enabled
}

func IsGCPAuthEnabled(gcp *podtypes.GCP) bool {
	return gcp != nil && gcp.Auth.ServiceAccount != ""
}
