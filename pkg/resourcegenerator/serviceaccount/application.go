package serviceaccount

import (
	"fmt"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/reconciliation"
	"github.com/kartverket/skiperator/pkg/util"
	"golang.org/x/exp/maps"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func generateForApplication(r reconciliation.Reconciliation) error {
	ctxLog := r.GetLogger()
	ctxLog.Debug("Attempting to generate service account for application", r.GetReconciliationObject().GetName())

	application, ok := r.GetReconciliationObject().(*skiperatorv1alpha1.Application)
	if !ok {
		return fmt.Errorf("failed to cast object to Application")
	}

	serviceAccount := corev1.ServiceAccount{ObjectMeta: metav1.ObjectMeta{Namespace: application.Namespace, Name: application.Name}}

	if util.IsCloudSqlProxyEnabled(application.Spec.GCP) {
		setCloudSqlAnnotations(&serviceAccount, application)
	}

	var obj client.Object = &serviceAccount
	r.AddResource(&obj)
	ctxLog.Debug("Finished generating service account for application", application.Name)
	return nil
}

func setCloudSqlAnnotations(serviceAccount *corev1.ServiceAccount, application *skiperatorv1alpha1.Application) {
	annotations := serviceAccount.GetAnnotations()
	if len(annotations) == 0 {
		annotations = make(map[string]string)
	}
	maps.Copy(annotations, map[string]string{
		"iam.gke.io/gcp-service-account": application.Spec.GCP.CloudSQLProxy.ServiceAccount,
	})
	serviceAccount.SetAnnotations(annotations)
}
