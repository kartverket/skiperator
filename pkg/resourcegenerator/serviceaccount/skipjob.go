package serviceaccount

import (
	"fmt"

	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/pkg/reconciliation"
	"github.com/kartverket/skiperator/pkg/util"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func init() {
	multiGenerator.Register(reconciliation.JobType, generateForSKIPJob)
}

func generateForSKIPJob(r reconciliation.Reconciliation) error {
	ctxLog := r.GetLogger()
	ctxLog.Debug("Attempting to generate service account for skipjob", "skipjob", r.GetSKIPObject().GetName())

	skipJob, ok := r.GetSKIPObject().(*skiperatorv1alpha1.SKIPJob)
	if !ok {
		return fmt.Errorf("failed to cast object to skipjob")
	}

	serviceAccount := corev1.ServiceAccount{ObjectMeta: metav1.ObjectMeta{Namespace: skipJob.Namespace, Name: skipJob.KindPostFixedName()}}

	if util.IsCloudSqlProxyEnabled(skipJob.Spec.Container.GCP) {
		setCloudSqlAnnotations(&serviceAccount, skipJob)
	}
	r.AddResource(&serviceAccount)
	ctxLog.Debug("Finished generating service account for skipjob", "skipjob", skipJob.Name)
	return nil
}
