package util

import (
	"context"
	"fmt"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/mitchellh/hashstructure/v2"
	"github.com/r3labs/diff/v3"
	"hash/fnv"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"reflect"
	"regexp"
	"strconv"
	"unicode"

	"golang.org/x/exp/maps"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var internalPattern = regexp.MustCompile(`[^.]\.skip\.statkart\.no`)

func IsInternal(hostname string) bool {
	return internalPattern.MatchString(hostname)
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

func IsHPAEnabled(jsonReplicas *apiextensionsv1.JSON) bool {
	replicas, err := skiperatorv1alpha1.GetReplicasStruct(jsonReplicas)
	if err == nil &&
		replicas.Min > 0 &&
		replicas.Min < replicas.Max {
		return true
	}
	return false
}

func GetConfigMap(client client.Client, ctx context.Context, namespacedName types.NamespacedName) (corev1.ConfigMap, error) {
	configMap := corev1.ConfigMap{}

	err := client.Get(ctx, namespacedName, &configMap)

	return configMap, err
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

func SetCommonAnnotations(object client.Object) {
	annotations := object.GetAnnotations()
	if len(annotations) == 0 {
		annotations = make(map[string]string)
	}
	maps.Copy(annotations, CommonAnnotations)
	object.SetAnnotations(annotations)
}

func PointTo[T any](x T) *T {
	return &x
}

func GetPodAppSelector(applicationName string) map[string]string {
	return map[string]string{"app": applicationName}
}

func HasUpperCaseLetter(word string) bool {
	for _, letter := range word {
		if unicode.IsUpper(letter) {
			return true
		}
	}

	return false
}

func ResourceNameWithHash(resourceName string, kind string) string {
	hash := GenerateHashFromName(resourceName + kind)

	return resourceName + "-" + strconv.FormatUint(hash, 16)
}

func GetObjectDiff[T any](a T, b T) {
	aKind := reflect.ValueOf(a).Kind()
	bKind := reflect.ValueOf(b).Kind()
	if aKind != bKind {
		fmt.Printf("The objects to compare are not the same, found obj1: %v, obj2: %v\n", aKind, bKind)
		return
	}
	changelog, _ := diff.Diff(a, b)

	if len(changelog) == 0 {
		fmt.Printf("No changes found\n")
	}

	fmt.Printf("Changes found: \n%v", changelog)
}
