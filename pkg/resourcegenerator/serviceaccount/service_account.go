package serviceaccount

import (
	"context"
	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/log"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/resourceutils"
	"github.com/kartverket/skiperator/pkg/util"
	"golang.org/x/exp/maps"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Generate(ctx context.Context, application *skiperatorv1alpha1.Application) *corev1.ServiceAccount {
	ctxLog := log.NewLogger(ctx)
	ctxLog.Debug("Attempting to generate service account for application", application.Name)

	serviceAccount := corev1.ServiceAccount{ObjectMeta: metav1.ObjectMeta{Namespace: application.Namespace, Name: application.Name}}

	resourceutils.SetApplicationLabels(&serviceAccount, application)
	resourceutils.SetCommonAnnotations(&serviceAccount)

	if util.IsCloudSqlProxyEnabled(application.Spec.GCP) {
		setCloudSqlAnnotations(&serviceAccount, application)
	}

	return &serviceAccount
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
