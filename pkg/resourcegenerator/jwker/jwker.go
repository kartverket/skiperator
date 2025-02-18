package jwker

import (
	"fmt"

	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/api/v1alpha1/podtypes"
	"github.com/kartverket/skiperator/pkg/reconciliation"
	"github.com/kartverket/skiperator/pkg/util"
	naisiov1 "github.com/nais/liberator/pkg/apis/nais.io/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Generate(r reconciliation.Reconciliation) error {
	ctxLog := r.GetLogger()
	if r.GetType() != reconciliation.ApplicationType {
		return fmt.Errorf("unsupported type %s in jwker resource", r.GetType())
	}

	application, ok := r.GetSKIPObject().(*skiperatorv1alpha1.Application)
	if !ok {
		err := fmt.Errorf("failed to cast resource to Application")
		ctxLog.Error(err, "failed to generate jwker resource")
		return err
	}

	if application.Spec.AccessPolicy == nil {
		return nil
	}

	if !application.Spec.AccessPolicy.TokenX {
		return nil
	}

	ctxLog.Debug("Attempting to generate jwker resource for application", "application", application.Name)

	var err error

	jwker := naisiov1.Jwker{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "nais.io/v1",
			Kind:       "Jwker",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: application.Namespace,
			Name:      application.Name,
		},
	}

	jwker.Spec, err = getJwkerSpec(application)
	if err != nil {
		return err
	}

	r.AddResource(&jwker)
	ctxLog.Debug("Finished generating jwker resource for application", "application", application.Name)

	return nil
}

func GetJwkerEnvVariables(secretName string) []corev1.EnvVar {
	variableNames := []string{"TOKEN_X_PRIVATE_JWK", "TOKEN_X_CLIENT_ID", "TOKEN_X_TOKEN_ENDPOINT", "TOKEN_X_JWKS_URI"}

	// and push jwker secrets to environment
	variables := []corev1.EnvVar{}
	for _, variableName := range variableNames {
		variables = append(variables, corev1.EnvVar{
			Name: variableName,
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					Key: variableName,
					LocalObjectReference: corev1.LocalObjectReference{
						Name: secretName,
					},
					Optional: util.PointTo(true),
				},
			},
		})
	}
	return variables
}

// Assumes application.Spec.AccessPolicy is not nil
func getJwkerSpec(application *skiperatorv1alpha1.Application) (naisiov1.JwkerSpec, error) {
	secretName, err := GetJwkerSecretName(application.Name)
	if err != nil {
		return naisiov1.JwkerSpec{}, err
	}

	jwkerAccessPolicy, err := GenerateJwkerAccessPolicy(application)
	if err != nil {
		return naisiov1.JwkerSpec{}, err
	}

	spec := naisiov1.JwkerSpec{
		AccessPolicy: jwkerAccessPolicy,
		SecretName:   secretName,
	}

	return spec, nil
}

func GenerateJwkerAccessPolicy(application *skiperatorv1alpha1.Application) (*naisiov1.AccessPolicy, error) {
	if application.Spec.AccessPolicy == nil {
		return nil, fmt.Errorf("AccessPolicy is nil")
	}

	accessPolicy := application.Spec.AccessPolicy.DeepCopy()

	jwkerAccessPolicy := &naisiov1.AccessPolicy{
		Inbound: &naisiov1.AccessPolicyInbound{
			Rules: []naisiov1.AccessPolicyInboundRule{},
		},
	}

	for _, rule := range accessPolicy.Inbound.Rules {
		jwkerRule := naisiov1.AccessPolicyInboundRule{
			AccessPolicyRule: naisiov1.AccessPolicyRule{
				Application: rule.Application,
				Namespace:   rule.Namespace,
			},
		}
		jwkerAccessPolicy.Inbound.Rules = append(jwkerAccessPolicy.Inbound.Rules, jwkerRule)
	}

	return jwkerAccessPolicy, nil
}

func GetJwkerSecretName(name string) (string, error) {
	return util.GetSecretName("jwker", name)
}

func TokenXSpecifiedInSpec(accessPolicy *podtypes.AccessPolicy) bool {
	return accessPolicy != nil && accessPolicy.TokenX
}
