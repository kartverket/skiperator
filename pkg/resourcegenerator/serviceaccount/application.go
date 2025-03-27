package serviceaccount

import (
	"fmt"

	skiperatorv1alpha1 "github.com/kartverket/skiperator/v3/api/v1alpha1"
	"github.com/kartverket/skiperator/v3/pkg/reconciliation"
	"github.com/kartverket/skiperator/v3/pkg/util"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func init() {
	multiGenerator.Register(reconciliation.ApplicationType, generateForApplication)
}

func generateForApplication(r reconciliation.Reconciliation) error {
	ctxLog := r.GetLogger()
	ctxLog.Debug("Attempting to generate service account for application", "application", r.GetSKIPObject().GetName())

	application, ok := r.GetSKIPObject().(*skiperatorv1alpha1.Application)
	if !ok {
		return fmt.Errorf("failed to cast object to Application")
	}

	serviceAccount := corev1.ServiceAccount{ObjectMeta: metav1.ObjectMeta{Namespace: application.Namespace, Name: application.Name}}

	if util.IsCloudSqlProxyEnabled(application.Spec.GCP) {
		setCloudSqlAnnotations(&serviceAccount, application)
	}
	r.AddResource(&serviceAccount)
	ctxLog.Debug("Finished generating service account for application", "application", application.Name)
	return nil
}
