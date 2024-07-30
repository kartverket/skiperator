package peerauthentication

import (
	"fmt"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/reconciliation"
	"github.com/kartverket/skiperator/pkg/util"
	securityv1beta1api "istio.io/api/security/v1beta1"
	typev1beta1 "istio.io/api/type/v1beta1"
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func Generate(r reconciliation.Reconciliation) error {
	ctxLog := r.GetLogger()
	if r.GetType() != reconciliation.ApplicationType {
		return fmt.Errorf("unsupported type %s in peer authentication", r.GetType())
	}
	application, ok := r.GetSKIPObject().(*skiperatorv1alpha1.Application)
	if !ok {
		err := fmt.Errorf("failed to cast resource to application")
		ctxLog.Error(err, "Failed to generate peer authentication")
		return err
	}
	ctxLog.Debug("Attempting to generate peer authentication for application", "application", application.Name)

	peerAuthentication := securityv1beta1.PeerAuthentication{ObjectMeta: metav1.ObjectMeta{Namespace: application.Namespace, Name: application.Name}}

	peerAuthentication.Spec = securityv1beta1api.PeerAuthentication{
		Selector: &typev1beta1.WorkloadSelector{
			MatchLabels: util.GetPodAppSelector(application.Name),
		},
		Mtls: &securityv1beta1api.PeerAuthentication_MutualTLS{
			Mode: securityv1beta1api.PeerAuthentication_MutualTLS_STRICT,
		},
	}

	ctxLog.Debug("Finished generating peer authentication for application", "application", application.Name)

	var obj client.Object = &peerAuthentication
	r.AddResource(&obj)

	return nil
}
