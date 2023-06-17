package util

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash"
	"hash/fnv"
	"regexp"
	"unicode"

	"golang.org/x/crypto/ripemd160"
	"golang.org/x/exp/maps"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var internalPattern = regexp.MustCompile(`[^.]\.skip\.statkart\.no`)

const HashLabelName = "skiperator.kartverket.no/hash"

func IsInternal(hostname string) bool {
	return internalPattern.MatchString(hostname)
}

func GetHashForSpec(specStruct interface{}) string {
	byteArray, _ := json.Marshal(specStruct)
	var hasher hash.Hash
	hasher = ripemd160.New()
	hasher.Reset()
	hasher.Write(byteArray)
	return hex.EncodeToString(hasher.Sum(nil))
}

func SetHashToLabels(labels map[string]string, specHashActual string) map[string]string {
	if labels == nil {
		labels = map[string]string{}
	}
	labels[HashLabelName] = specHashActual
	return labels
}

func GetHashFromLabels(labels map[string]string) string {
	return labels[HashLabelName]
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

func GetApplicationSelector(applicationName string) map[string]string {
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
